import hashlib
from base64 import b64encode
import requests
import time

coin_prefix = 'CPEN 442 Coin' + '2021'
public_id = '6a227691f429c72bbcde8face94b48c3b505fb8b7de7371cc403bfa383b48e3f'
poll_interval = 60 * 20


def assemble_bytes(hash_of_preceding_coin, coin_blob, id_of_miner):
    return coin_prefix.encode('ascii') + hash_of_preceding_coin.encode('ascii') + coin_blob.to_bytes((coin_blob.bit_length() + 7) // 8, 'big') + id_of_miner.encode('ascii')

def get_difficulty():
    url = 'http://cpen442coin.ece.ubc.ca/difficulty'
    res = requests.post(url)
    if res:
        diff = res.json()['number_of_leading_zeros']
        print('difficulty_level:', diff, res.json())
        return diff
    else:
        print('Error fetching difficulty.', res.json())
        return 4

def get_prev_hash():
    url = 'http://cpen442coin.ece.ubc.ca/last_coin'
    res = requests.post(url)
    if res:
        prev_hash = res.json()['coin_id']
        print('hash_of_preceding_coin:', prev_hash, res.json())
        return prev_hash
    else:
        print('Error fetching previous hash.', res.json())
        return ''

def check_coin(hexstring, difficulty_level):
    return hexstring[:difficulty_level] == '0' * difficulty_level

def claim_coin(coin_blob):
    url = 'http://cpen442coin.ece.ubc.ca/claim_coin'
    encoded_coin = b64encode(coin_blob.to_bytes((coin_blob.bit_length() + 7) // 8, 'big')).decode()
    data = {
        'coin_blob': encoded_coin,
        'id_of_miner': public_id
    }
    res = requests.post(url, json=data)
    if res:
        print('claiming coin: SUCCESS :D', res.json())
    else:
        print('claiming coin: FAILED :\'(', res.json())
    

def mine_coins():
    difficulty_level = get_difficulty()
    hash_of_preceding_coin = get_prev_hash()
    last_poll_time = time.perf_counter()
    coin_blob = 0
    
    # Start mining coin
    while True:
        bytestr = assemble_bytes(hash_of_preceding_coin, coin_blob, public_id)
        h = hashlib.sha256(bytestr)
        hash_hexstring = h.hexdigest()

        if check_coin(hash_hexstring, difficulty_level):
            print('wooo!')
            print('\t', coin_blob, hex(coin_blob))
            print('\t', hex(int.from_bytes(bytestr, 'big')))
            print('\t', hash_hexstring)

            # Claim coin
            claim_coin(coin_blob)
            
            # Update params
            difficulty_level = get_difficulty()
            poll_prev_hash = get_prev_hash()
            last_poll_time = time.perf_counter()
            
            # If prev hash changed, start over
            if (poll_prev_hash != hash_of_preceding_coin):
                hash_of_preceding_coin = poll_prev_hash
                coin_blob = 0
                continue
        
        coin_blob += 1

        curr_time = time.perf_counter()
        if (curr_time - last_poll_time >= poll_interval):
            # Update params
            difficulty_level = get_difficulty()
            poll_prev_hash = get_prev_hash()
            last_poll_time = curr_time

            # If prev hash changed, start over
            if (poll_prev_hash != hash_of_preceding_coin):
                hash_of_preceding_coin = poll_prev_hash
                coin_blob = 0
                continue


mine_coins()
