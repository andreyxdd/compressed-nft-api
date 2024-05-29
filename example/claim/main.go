package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"

	"github.com/xssnick/tonutils-go/address"
	"github.com/xssnick/tonutils-go/tvm/cell"
)

const COMPRESSED_NFT_API_URL = "http://localhost:8081/v1"

func main() {
	// 1. Fetch collection state
	resp, err := http.Get(fmt.Sprintf("%s/state", COMPRESSED_NFT_API_URL))
	if err != nil {
		log.Fatalln("Fetch state err:", err.Error())
		return
	}
	defer resp.Body.Close()

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalln("ReadAll state err:", err.Error())
		return
	}

	var state struct {
		Address string `json:"address"`
	}
	err = json.Unmarshal(b, &state)
	if err != nil {
		log.Fatalln("Unmarshal state err:", err.Error())
		return
	}

	// 2. Fetch proof cell for a specific item (by its index)
	var itemIndex uint64 = 1
	resp, err = http.Get(fmt.Sprintf("%s/items/%d", COMPRESSED_NFT_API_URL, itemIndex))
	if err != nil {
		log.Fatalln("Fetch item err:", err.Error())
		return
	}
	defer resp.Body.Close()

	b, err = io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalln("ReadAll item err:", err.Error())
		return
	}

	var item struct {
		ProofCellBase64 string `json:"proof_cell"`
	}
	err = json.Unmarshal(b, &item)
	if err != nil {
		log.Fatalln("Unmarshal item err:", err.Error())
		return
	}

	// 3. Prepare payload for the claim transacation
	proofCellDecoded, err := base64.StdEncoding.DecodeString(item.ProofCellBase64)
	if err != nil {
		log.Fatalln("DecodeString err:", err.Error())
		return
	}
	proofCell, err := cell.FromBOC(proofCellDecoded)
	if err != nil {
		log.Fatalln("FromBOC err:", err.Error())
		return
	}
	payload := cell.BeginCell().
		MustStoreUInt(0x13a3ca6, 32).
		MustStoreUInt(rand.Uint64(), 64).
		MustStoreUInt(itemIndex, 256).
		MustStoreRef(proofCell).
		EndCell()

	// 4. Complete the transaction
	collectionAddrs := address.MustParseRawAddr(state.Address)
	amount := "85000000" // Claim transaction min is 0.085 TON (https://github.com/ton-community/compressed-nft-contract/blob/de529e67971f0c888ef496cf983d681a5a13a06b/contracts/collection_new.fc#L22)
	link := fmt.Sprintf(
		"ton://transfer/%v?amount=%s&bin=%v",
		collectionAddrs.String(),
		amount,
		base64.RawURLEncoding.EncodeToString(payload.ToBOC()),
	)
	fmt.Println(link)
}
