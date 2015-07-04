## HashChash(-like) resource control

To limit posting and retaining of messages the repserver requires each message
to contain a hashcash-like proof of work. Contrary to the standard hashcash
scheme, the scheme used by repbin allows the pre- calculation and reuse of
proofs.  With repbin, the sending party generates a new ed25519 keypair and
calculates the hash collision for the public key. It then uses the public key to
signs the complete message with the ed25519 and attaches the hashcash nonce. The
repserver keeps track of the number of messages posted/retained using one
ed25519 key and calculates limits accordingly. The client can continue with the
hashcollision after the key was used first, potentially adding more collision
bits and thus more resources of the server. The client can also pre-calculate
ed25519 keys and the respective hash collision to speed up operation in case of
post. For privacy reasons, a client SHOULD use fresh ed25519/hashcash keys for
each post. This is the default behavior of the client.


## Long-Term recipient key attributes

The first two bits of the recipient key determine if a message to that key will
be replicated to other peers by adding or omitting it from the global message
index of the repserver it was posted to. They also control if the Post-Box for
this key is publicly accessible or not. Messages to a key that is marked as
hidden (post-box access requires authentication) are still replicated unless the
no-sync bit is also set. The enforcement of both attributes relies exclusively
on the behavior of the repserver and is not cryptographically secured.


## Post-Box access control

For every recipient key the repserver keeps a list of messages that were sent to
that key. Access to that list can be controlled by the key attributes. If a key
is marked as "hidden", access to the Post-Box requires authentication. This
authentication is done by a proof-of-knowledge for the private key of the
recipient public key. The cryptography for that protocol is implemented in
utils/keyauth. The client first needs to fetch time-dependent curve25519 public
key from the repserver. It then calculates a diffie-hellman shared secret
between the repserver key and his private key. The shared secret is then hashed
with SHA256. When accessing the post-box, the has of the shared secret is
verified by the repserver and on success, access is granted. Please be aware
that the temporary public keys of the repserver must be considered insecure for
any task but authentication and may not be used for any other operation.
Authentication data is valid for a limited time (usually a few minutes) and need
to be re-calculated for each post-box access.


## Peer authentication

Peers mutually authenticate each other with a protocol based on ed25519
signatures. On notification, Peer A sends a temporary authentication token to
Peer B. The token contains the time, Peer A's public key, Peer B's public key
and a signature by Peer A. Peer B verifies the token as soon as he receives it
and signs it with his own public key. The signed token is send to Peer A when
the global index needs to be accessed and is verified, including Peer B's
signature, by Peer A. Authentication tokens are valid for a limited time
(usually 2 days) and are renewed on each notification.


## Message format and message cryptography

Messages are base64 encoded. They contain the following sections and fields (in
order):

```
SignatureHeader:
	Version:    1 byte = 0x01
	PubkeySign: ed25519 public key
	PubkeyCash: HashCash Nonce (8 byte)
	Signature:  ed25519 signature over MessageID
	MessageID   hash256(Envelope)
Envelope:
	KeyHeader:
		PubKeySendConstant     curve25519 pubkey
		PubKeySendTemporary    curve25519 pubkey
		PubKeyReceiveConstant  curve25519 pubkey
		PubKeyReceiveTemporary curve25519 pubkey
		Nonce 32byte
	EncryptedBody
	HMAC-SHA256 (EncryptedBody)
```

The SignatureHeader is used exclusively for the repserver protocol.
The KeyHeader contains information to calculate the shared secret that is used
for the encryption of Body and the HMAC of the encrypted body (encrypt then
hmac).

The sender calculates the shared secret with the following operation:

```
	PreKey_1 := DH-CURVE25519(PrivKeySendTemporary, PubKeyReceiveTemporary)
	PreKey_2 := DH-CURVE25519(PrivKeySendConstant,  PubKeyReceiveTemporary)
	PreKey_3 := DH-CURVE25519(PrivKeySendTemporary, PubKeyReceiveConstant )
	PreKey   := hash512(PreKey_1 | PreKey_2 | PreKey_3 | Nonce)
(The recipient swaps PreKey_2 and PreKey_3. "|" denotes concatenation)
```

In default operation, all keys are ephemeral. In this case, the PubKeyReceiveConstant determines PubKeyReceiveTemporary by:

```
	DETERMGEN := Key fingerprint of "Replicated Encrypted Pastebin" 0xDCA2BA86 OpenPGP key (repeated to 32 byte)
	PubKeyReceiveTemporary-Private := SHA256(PubKeyReceiveConstant-Private | DETERMGEN)
```

For all cryptographic operations requiring random bits the operating system
random source is used directly.

From the shared secret two keys are derived using two generators and HMAC-SHA256
```
	HMACKey      := HMAC-SHA256(HMACGEN, SharedSecret)
	SymmetricKey := HMAC-SHA256(SYMMGEN, SharedSecret)
```

The generators are:
```
	HMACGEN := SHA256("Repbin HMAC Key")
	SYMMGEN := SHA256("Repbin Symmetric Key")
```

The HMAC of the EncryptedBody is calculated as
```
	HMACBody := HMAC-SHA256(body, HMACKey)
```

The Body is encrypted using AES256-CTR. The key is SymmetricKey, the IV is the
SHA256 of the KeyHeader including Nonce. The counter starts with zero.

The cleartext body contains additional fields:
```
	MessageType: 1 byte (0x01 = Standard Message, 0x02 = List of messages, 0x03 Repost message )
	ContentLength: 2 bytes. unsigned big endian
	Content:
		ReplyPubKeyConstant  curve25519 pubkey (or all zeros)
		ReplyPubKeyTemporary curve25519 pubkey (or all zeros)
		Data (MessageType specific)
	RandomPadding
	DeterministicPadding
```

All padding is generated by expanding a random 256 bit value with AES256-CTR (IV
and Counter are zero). The length of both paddings combined fills the message to
the maximum message length. Padding does flow into the HMAC and MessageID. The
keys for padding generation are random and they are deleted as soon as the
padding has been generated. A special case exists for repost messages.

### Repost

Repost messages contain other messages with the DeterministicPadding removed
(but all other fields present). On repost, the PaddingKey is read from the body
to generate the DeterministicPadding and insert it into the embeded message. The
embedded message is then posted honoring the MinDelay and MaxDelay settings.
For a repost message, the Data section looks like this:

```
	Data:
		MinDelay 8 byte unsigned big endian
		MaxDelay 8 byte unsigned big endian
		PaddingKey 32 byte
		EmbeddedMessage
```

Key generation and all cryptographic operations can be observed by calling the
client with the --KEYVERB parameter. This will print the private and temporary
keys to stderr.
