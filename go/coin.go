package main

import (
    "encoding/binary"
    "crypto/sha256"
	b64 "encoding/base64"
    "bytes"
    "encoding/json"
    "io/ioutil"
    "net/http"
    "strings"
    "time"
    "fmt"
)

const coin_prefix = "CPEN 442 Coin" + "2021"
const public_id = "6a227691f429c72bbcde8face94b48c3b505fb8b7de7371cc403bfa383b48e3f"
const poll_interval = 60 * 20

func assemble_bytes(hash_of_preceding_coin string, coin_blob uint32, id_of_miner string) []byte {
    coin_blob_bytes := make([]byte, binary.Size(coin_blob))
    binary.BigEndian.PutUint32(coin_blob_bytes, coin_blob)

    // fmt.Printf("%x\n", []byte(coin_prefix))
    // fmt.Printf("%x\n", []byte(hash_of_preceding_coin))
    // fmt.Printf("%x\n", coin_blob_bytes)
    // fmt.Printf("%x\n", []byte(id_of_miner))
    
    return bytes.Join([][]byte{[]byte(coin_prefix), []byte(hash_of_preceding_coin), coin_blob_bytes, []byte(id_of_miner)}, nil)
}

func hash(bytestr []byte) string {
    hash_bytes := sha256.Sum256(bytestr)
    return string(hash_bytes[:])
}

func get_difficulty() int {
    url := "http://cpen442coin.ece.ubc.ca/difficulty"
    res, err := http.Post(url, "application/json", nil)
    if err != nil {
        fmt.Println("Error fetching difficulty.")
        return 4
    }
    defer res.Body.Close()
    // body, err := ioutil.ReadAll(res.Body)
    // fmt.Printf("response body: %s\n", string(body))
    var result map[string] interface{}
    json.NewDecoder(res.Body).Decode(&result)
    // fmt.Println(result)
    diff := int(result["number_of_leading_zeros"].(float64))
    fmt.Printf("difficulty_level: %d\n", diff)
    return diff
}

func get_prev_hash() string {
    url := "http://cpen442coin.ece.ubc.ca/last_coin"
    res, err := http.Post(url, "application/json", nil)
    if err != nil {
        fmt.Println("Error fetching previous hash.")
        return ""
    }
    defer res.Body.Close()
    var result map[string] interface{}
    json.NewDecoder(res.Body).Decode(&result)
    prev_hash := result["coin_id"].(string)
    fmt.Printf("hash_of_preceding_coin: %s\n", prev_hash)
    return prev_hash
}

func check_coin(hexstring string, difficulty_level int) bool {
    return hexstring[:difficulty_level] == strings.Repeat("0", difficulty_level)
}

func claim_coin(coin_blob uint32) {
    url := "http://cpen442coin.ece.ubc.ca/claim_coin"
    coin_blob_bytes := make([]byte, binary.Size(coin_blob))
    binary.BigEndian.PutUint32(coin_blob_bytes, coin_blob)
    encoded_coin := b64.StdEncoding.EncodeToString(coin_blob_bytes)
    data, _ := json.Marshal(map[string]string{
        "coin_blob": encoded_coin,
        "id_of_miner": public_id,
    })
    res, _ := http.Post(url, "application/json", bytes.NewBuffer(data))
    body, _ := ioutil.ReadAll(res.Body)
    fmt.Printf("claiming coin: %s\n", string(body))
}

func mine_coins() {
    difficulty_level := get_difficulty()
    hash_of_preceding_coin := get_prev_hash()
    last_poll_time := time.Now()
    var coin_blob uint32 = 0
    var curr_time time.Time
    var poll_prev_hash string
    
    // Start mining coin
    for {
        bytestr := assemble_bytes(hash_of_preceding_coin, coin_blob, public_id)
        hash_string := hash(bytestr)

        if check_coin(hash_string, difficulty_level) {
            fmt.Printf("wooo! %d %s\n", coin_blob, hash_string)

            // Claim coin
            claim_coin(coin_blob)
            
            // Update params
            difficulty_level = get_difficulty()
            poll_prev_hash = get_prev_hash()
            last_poll_time = time.Now()
            
            // If prev hash changed, start over
            if (poll_prev_hash != hash_of_preceding_coin) {
                hash_of_preceding_coin = poll_prev_hash
                coin_blob = 0
                continue
            }
        }
        
        coin_blob += 1

        curr_time = time.Now()
        if (curr_time.Sub(last_poll_time).Seconds() >= poll_interval) {
            // Update params
            difficulty_level = get_difficulty()
            poll_prev_hash = get_prev_hash()
            last_poll_time = curr_time

            // If prev hash changed, start over
            if (poll_prev_hash != hash_of_preceding_coin) {
                hash_of_preceding_coin = poll_prev_hash
                coin_blob = 0
                continue
            }
        }
    }
}

func test() {
    bytestr := assemble_bytes("a9c1ae3f4fc29d0be9113a42090a5ef9fdef93f5ec4777a008873972e60bb532", 2142603169, "5e18025278a10c33c32d441b211db2d75f43c58898a88ce94b43552eafe17da0")
    fmt.Printf("assemble_bytes: %x\n", bytestr)
    fmt.Printf("hash: %x\n", hash(bytestr))
    get_difficulty()
    get_prev_hash()
    fmt.Printf("check_coin: %v\n", check_coin("00beef", 4))
    fmt.Printf("check_coin: %v\n", check_coin("0000aa", 4))
    fmt.Printf("check_coin: %v\n", check_coin("00000a", 4))
    claim_coin(123456)
}

func main() {
    mine_coins()
    // test()
}
