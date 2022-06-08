// Copyright (c) 2022 The illium developers
// Use of this source code is governed by an MIT
// license that can be found in the LICENSE file.

package harness

import (
	"crypto/rand"
	"github.com/libp2p/go-libp2p-core/crypto"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/project-illium/ilxd/blockchain"
	"github.com/project-illium/ilxd/params"
	"github.com/project-illium/ilxd/types"
	"github.com/project-illium/ilxd/types/blocks"
	"github.com/project-illium/ilxd/types/transactions"
	"github.com/project-illium/ilxd/wallet"
	"github.com/project-illium/ilxd/zk"
	"github.com/project-illium/ilxd/zk/circuits/stake"
	"github.com/project-illium/ilxd/zk/circuits/standard"
	"time"
)

func (h *TestHarness) generateBlocks(nBlocks int) ([]*blocks.Block, error) {
	newBlocks := make([]*blocks.Block, 0, nBlocks)
	acc := h.acc.Clone()
	fee := uint64(1)

	for n := 0; n < nBlocks; n++ {
		outputsPerTx := h.txsPerBlock
		numTxs := h.txsPerBlock
		if len(h.spendableNotes) < h.txsPerBlock {
			outputsPerTx = h.txsPerBlock / len(h.spendableNotes)
			numTxs = len(h.spendableNotes)
		}

		notes := make([]*SpendableNote, 0, len(h.spendableNotes))
		for _, note := range h.spendableNotes {
			notes = append(notes, note)
		}

		txs := make([]*transactions.Transaction, 0, len(h.spendableNotes))
		for i := 0; i < numTxs; i++ {
			sn := notes[i]

			commitment, err := sn.Note.Commitment()
			if err != nil {
				return nil, err
			}
			inclusionProof, err := acc.GetProof(commitment)
			if err != nil {
				return nil, err
			}
			serializedKey, err := sn.Note.SpendScript.Pubkeys[0].Serialize()
			if err != nil {
				return nil, err
			}

			nullifier, err := types.CalculateNullifier(inclusionProof.Index, sn.Note.Salt, sn.Note.SpendScript.Threshold, serializedKey)
			if err != nil {
				return nil, err
			}

			var (
				outputs           = make([]*transactions.Output, 0, outputsPerTx)
				outputCommitments = make([][]byte, outputsPerTx)
				outputNotes       = make([]*SpendableNote, outputsPerTx)
			)

			for x := 0; x < outputsPerTx; x++ {
				privKey, _, err := crypto.GenerateEd25519Key(rand.Reader)
				if err != nil {
					return nil, err
				}

				var salt [32]byte
				rand.Read(salt[:])

				outputNote := &wallet.SpendNote{
					SpendScript: wallet.SpendScript{
						Threshold: 1,
						Pubkeys:   []*wallet.TimeLockedPubkey{wallet.NewTimeLockedPubkey(privKey.GetPublic(), time.Time{})},
					},
					Amount:  (sn.Note.Amount / uint64(outputsPerTx)) - fee,
					AssetID: wallet.IlliumCoinID,
					Salt:    salt,
				}
				outputNotes[i] = &SpendableNote{
					Note:       outputNote,
					PrivateKey: privKey,
				}

				outputCommitment, err := outputNote.Commitment()
				if err != nil {
					return nil, err
				}
				outputCommitments[i] = outputCommitment

				outputs = append(outputs, &transactions.Output{
					Commitment:      outputCommitment,
					EphemeralPubkey: make([]byte, blockchain.PubkeyLen),
					Ciphertext:      make([]byte, blockchain.CipherTextLen),
				})
			}
			standardTx := &transactions.StandardTransaction{
				Outputs:    outputs,
				Fee:        1,
				Nullifiers: [][]byte{nullifier.Bytes()},
				TxoRoot:    acc.Root().Bytes(),
				Proof:      nil,
			}

			sigHash, err := standardTx.SigHash()
			if err != nil {
				return nil, err
			}
			sig, err := sn.PrivateKey.Sign(sigHash)
			if err != nil {
				return nil, err
			}

			privateParams := &standard.PrivateParams{
				Inputs: []standard.PrivateInput{
					{
						Amount:          sn.Note.Amount,
						Salt:            sn.Note.Salt[:],
						AssetID:         sn.Note.AssetID,
						CommitmentIndex: inclusionProof.Index,
						InclusionProof: standard.InclusionProof{
							Hashes:      inclusionProof.Hashes,
							Flags:       inclusionProof.Flags,
							Accumulator: inclusionProof.Accumulator,
						},
						Threshold:   1,
						Pubkeys:     [][]byte{serializedKey},
						Signatures:  [][]byte{sig},
						SigBitfield: 1,
					},
				},
			}
			for _, outNote := range outputNotes {
				spendScript, err := outNote.Note.SpendScript.Hash()
				if err != nil {
					return nil, err
				}
				privateParams.Outputs = append(privateParams.Outputs, standard.PrivateOutput{
					SpendScript: spendScript,
					Amount:      outNote.Note.Amount,
					Salt:        outNote.Note.Salt[:],
					AssetID:     outNote.Note.AssetID,
				})
			}

			publicPrams := &standard.PublicParams{
				TXORoot:           h.acc.Root().Bytes(),
				SigHash:           sigHash,
				OutputCommitments: outputCommitments,
				Nullifiers:        [][]byte{nullifier.Bytes()},
				Fee:               fee,
				Blocktime:         time.Now(),
			}

			proof, err := zk.CreateSnark(standard.StandardCircuit, privateParams, publicPrams)
			if err != nil {
				return nil, err
			}
			standardTx.Proof = proof
			txs = append(txs, transactions.WrapTransaction(standardTx))
		}

		bestID, bestHeight := h.chain.BestBlock()

		txids := make([][]byte, 0, len(txs))
		for _, tx := range txs {
			txids = append(txids, tx.ID().Bytes())
		}
		merkles := blockchain.BuildMerkleTreeStore(txids)

		h.timeSource++

		var (
			networkKey crypto.PrivKey
			validator  peer.ID
		)
		for k, v := range h.validators {
			networkKey = v.networkKey
			validator = k
		}
		valBytes, err := validator.Marshal()
		if err != nil {
			return nil, err
		}

		blk := &blocks.Block{
			Header: &blocks.BlockHeader{
				Version:     1,
				Height:      bestHeight + 1,
				Parent:      bestID.Bytes(),
				Timestamp:   h.timeSource,
				TxRoot:      merkles[len(merkles)-1],
				Producer_ID: valBytes,
				Signature:   nil,
			},
			Transactions: txs,
		}

		sigHash, err := blk.Header.SigHash()
		if err != nil {
			return nil, err
		}
		sig, err := networkKey.Sign(sigHash)
		if err != nil {
			return nil, err
		}
		blk.Header.Signature = sig

		newBlocks = append(newBlocks, blk)
		for _, out := range blk.Outputs() {
			acc.Insert(out.Commitment, true)
		}
	}
	return newBlocks, nil
}

func createGenesisBlock(params *params.NetworkParams, networkKey, spendKey crypto.PrivKey,
	initialCoins uint64, additionalOutputs []*transactions.Output) (*blocks.Block, *wallet.SpendNote, error) {

	// First we'll create the spend note for the coinbase transaction.
	// The initial coins will be generated to the spendKey.
	var salt1 [32]byte
	rand.Read(salt1[:])

	note1 := &wallet.SpendNote{
		SpendScript: wallet.SpendScript{
			Threshold: 1,
			Pubkeys:   []*wallet.TimeLockedPubkey{wallet.NewTimeLockedPubkey(spendKey.GetPublic(), time.Time{})},
		},
		Amount:  initialCoins / 2,
		AssetID: wallet.IlliumCoinID,
		Salt:    salt1,
	}

	var salt2 [32]byte
	rand.Read(salt2[:])

	note2 := &wallet.SpendNote{
		SpendScript: wallet.SpendScript{
			Threshold: 1,
			Pubkeys:   []*wallet.TimeLockedPubkey{wallet.NewTimeLockedPubkey(spendKey.GetPublic(), time.Time{})},
		},
		Amount:  initialCoins / 2,
		AssetID: wallet.IlliumCoinID,
		Salt:    salt2,
	}

	// Next we're going to start building the coinbase transaction
	commitment1, err := note1.Commitment()
	if err != nil {
		return nil, nil, err
	}
	commitment2, err := note2.Commitment()
	if err != nil {
		return nil, nil, err
	}
	validatorID, err := peer.IDFromPublicKey(networkKey.GetPublic())
	if err != nil {
		return nil, nil, err
	}
	idBytes, err := validatorID.Marshal()
	if err != nil {
		return nil, nil, err
	}

	coinbaseTx := &transactions.CoinbaseTransaction{
		Validator_ID: idBytes,
		NewCoins:     initialCoins,
		Outputs: []*transactions.Output{
			{
				Commitment:      commitment1,
				EphemeralPubkey: make([]byte, blockchain.PubkeyLen),
				Ciphertext:      make([]byte, blockchain.CipherTextLen),
			},
			{
				Commitment:      commitment2,
				EphemeralPubkey: make([]byte, blockchain.PubkeyLen),
				Ciphertext:      make([]byte, blockchain.CipherTextLen),
			},
		},
	}
	coinbaseTx.Outputs = append(coinbaseTx.Outputs, additionalOutputs...)

	// And now sign the coinbase transaction with the network key
	sigHash, err := coinbaseTx.SigHash()
	if err != nil {
		return nil, nil, err
	}

	sig, err := networkKey.Sign(sigHash)
	if err != nil {
		return nil, nil, err
	}
	coinbaseTx.Signature = sig

	// Finally we're going to create the zk-snark proof for the coinbase
	// transaction.
	spendScriptHash1, err := note1.SpendScript.Hash()
	if err != nil {
		return nil, nil, err
	}

	spendScriptHash2, err := note2.SpendScript.Hash()
	if err != nil {
		return nil, nil, err
	}

	serializedPubkey1, err := note1.SpendScript.Pubkeys[0].Serialize()
	if err != nil {
		return nil, nil, err
	}

	serializedPubkey2, err := note1.SpendScript.Pubkeys[0].Serialize()
	if err != nil {
		return nil, nil, err
	}

	nullifier1, err := types.CalculateNullifier(0, salt1, 1, serializedPubkey1)
	if err != nil {
		return nil, nil, err
	}
	nullifier2, err := types.CalculateNullifier(1, salt2, 1, serializedPubkey2)
	if err != nil {
		return nil, nil, err
	}

	publicParams := &standard.PublicParams{
		OutputCommitments: [][]byte{commitment1, commitment2},
		Nullifiers:        [][]byte{nullifier1.Bytes(), nullifier2.Bytes()},
		Fee:               0,
		Coinbase:          initialCoins,
	}
	privateParams := &standard.PrivateParams{
		Outputs: []standard.PrivateOutput{
			{
				SpendScript: spendScriptHash1,
				Amount:      initialCoins / 2,
				Salt:        note1.Salt[:],
				AssetID:     note1.AssetID,
			},
			{
				SpendScript: spendScriptHash2,
				Amount:      initialCoins / 2,
				Salt:        note2.Salt[:],
				AssetID:     note2.AssetID,
			},
		},
	}

	proof, err := zk.CreateSnark(standard.StandardCircuit, privateParams, publicParams)
	if err != nil {
		return nil, nil, err
	}
	coinbaseTx.Proof = proof

	// Next we have to build the transaction staking the coins generated
	// in the prior coinbase transaction. This is needed because if no
	// validators are set in the genesis block we can't move the chain
	// forward.
	//
	// Notice there is a special validation rule for the genesis block
	// that doesn't apply to any other block. Normally, transactions
	// must contain a txoRoot for a block already in the chain. However,
	// in the case of the genesis block there are no other blocks in the
	// chain yet. So the rules allow the genesis block to reference its
	// own txoRoot.
	acc := blockchain.NewAccumulator()
	for i, output := range coinbaseTx.Outputs {
		acc.Insert(output.Commitment, i == 0)
	}
	txoRoot := acc.Root()
	inclusionProof, err := acc.GetProof(commitment1)
	if err != nil {
		return nil, nil, err
	}

	stakeTx := &transactions.StakeTransaction{
		Validator_ID: idBytes,
		Amount:       initialCoins,
		Nullifier:    nullifier1.Bytes(),
		TxoRoot:      txoRoot.Bytes(), // See note above
	}

	// Sign the stake transaction
	sigHash2, err := stakeTx.SigHash()
	if err != nil {
		return nil, nil, err
	}

	sig2, err := networkKey.Sign(sigHash2)
	if err != nil {
		return nil, nil, err
	}
	stakeTx.Signature = sig2

	// And generate the zk-snark proof
	sig3, err := spendKey.Sign(sigHash2)
	if err != nil {
		return nil, nil, err
	}

	publicParams2 := &stake.PublicParams{
		TXORoot:   txoRoot.Bytes(),
		SigHash:   sigHash2,
		Amount:    initialCoins / 2,
		Nullifier: nullifier1.Bytes(),
	}
	privateParams2 := &stake.PrivateParams{
		AssetID:         wallet.IlliumCoinID[:],
		Salt:            salt1[:],
		CommitmentIndex: 0,
		InclusionProof: standard.InclusionProof{
			Hashes:      inclusionProof.Hashes,
			Flags:       inclusionProof.Flags,
			Accumulator: inclusionProof.Accumulator,
		},
		Threshold:   1,
		Pubkeys:     [][]byte{serializedPubkey1},
		Signatures:  [][]byte{sig3},
		SigBitfield: 1,
	}

	proof2, err := zk.CreateSnark(stake.StakeCircuit, privateParams2, publicParams2)
	if err != nil {
		return nil, nil, err
	}
	stakeTx.Proof = proof2

	// Now we add the transactions to the genesis block
	genesis := params.GenesisBlock
	genesis.Transactions = []*transactions.Transaction{
		transactions.WrapTransaction(coinbaseTx),
		transactions.WrapTransaction(stakeTx),
	}

	// And create the genesis merkle root
	txids := make([][]byte, 0, len(genesis.Transactions))
	for _, tx := range genesis.Transactions {
		txids = append(txids, tx.ID().Bytes())
	}
	merkles := blockchain.BuildMerkleTreeStore(txids)
	genesis.Header.TxRoot = merkles[len(merkles)-1]
	genesis.Header.Timestamp = time.Now().Unix()

	return &genesis, note2, nil
}
