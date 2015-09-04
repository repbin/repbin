package sql

import (
	"strconv"

	"github.com/agl/ed25519"
	"github.com/repbin/repbin/hashcash"
	"github.com/repbin/repbin/message"
)

var (
	queries = map[string]map[string]string{
		"mysql": map[string]string{
			"SignerCreate": `CREATE TABLE IF NOT EXISTS signer (
                    ID BIGINT UNSIGNED NOT NULL PRIMARY KEY AUTO_INCREMENT,
                    PublicKey VARCHAR(` + strconv.FormatInt(message.SignerPubKeySize*2, 10) + `) NOT NULL,
                    Nonce VARCHAR(` + strconv.FormatInt(hashcash.NonceSize*2, 10) + `) NOT NULL,
                    Bits INT NOT NULL DEFAULT 0,
                    MessagesPosted BIGINT NOT NULL DEFAULT 0,
                    MessagesRetained BIGINT NOT NULL DEFAULT 0,
                    MaxMessagesPosted BIGINT UNSIGNED NOT NULL DEFAULT 0,
                    MaxMessagesRetained BIGINT UNSIGNED NOT NULL DEFAULT 0,
                    ExpireTarget BIGINT UNSIGNED NOT NULL DEFAULT 0,
                    LastMessageDeleted BIGINT UNSIGNED NOT NULL DEFAULT 0,
                    UNIQUE KEY pk (PublicKey)
                );`,
			"SignerInsert": `INSERT INTO signer (PublicKey,Nonce,Bits,
                    MaxMessagesPosted,MaxMessagesRetained,ExpireTarget) VALUES
                    (?,?,?,?,?,?)
                ;`,
			"SelectSignerPublicKey": `SELECT ID, PublicKey, Nonce, Bits, MessagesPosted,
                    MessagesRetained, MaxMessagesPosted, MaxMessagesRetained, ExpireTarget FROM
                    signer WHERE PublicKey=?
                ;`,
			"SelectSignerID": `SELECT ID, PublicKey, Nonce, Bits, MessagesPosted,
                    MessagesRetained, MaxMessagesPosted, MaxMessagesRetained, ExpireTarget FROM
                    signer WHERE ID=?
                ;`,
			"UpdateSigner": `UPDATE signer SET Nonce=?, Bits=?, MaxMessagesPosted=?, 
                    MaxMessagesRetained=?, ExpireTarget=? WHERE PublicKey=?
                ;`,
			"UpdateOrInsertSigner": `INSERT INTO signer 
                    (PublicKey, Nonce, Bits, MaxMessagesPosted, MaxMessagesRetained, ExpireTarget)
                    VALUES (?, ?, ?, ?, ?, ?) ON DUPLICATE KEY UPDATE
                    Nonce=VALUES(Nonce),
                    Bits=VALUES(Bits),
                    MaxMessagesPosted=VALUES(MaxMessagesPosted),
                    MaxMessagesRetained=VALUES(MaxMessagesRetained),
                    ExpireTarget=VALUES(ExpireTarget)
                ;`,
			"AddMessageSigner": `UPDATE signer 
                SET MessagesPosted=MessagesPosted+1, MessagesRetained=MessagesRetained+1, LastMessageDeleted=0
                WHERE PublicKey=? AND MaxMessagesRetained>MessagesRetained AND MaxMessagesPosted>MessagesPosted
                ;`,
			"DelMessageSigner": `UPDATE signer
                SET MessagesRetained=MessagesRetained-1
                WHERE PublicKey=? 
                ;`,
			"PrepareExpireSigner": `UPDATE signer SET LastMessageDeleted=? WHERE 
                MessagesRetained<=0 AND LastMessageDeleted=0
                ;`,
			"DeleteExpireSigner": `DELETE FROM signer 
                WHERE LastMessageDeleted!=0 AND LastMessageDeleted<?
                ;`,
			"PeerCreate": `CREATE TABLE IF NOT EXISTS peer (
                    ID BIGINT UNSIGNED NOT NULL PRIMARY KEY AUTO_INCREMENT,
                    PublicKey VARCHAR(` + strconv.FormatInt(ed25519.PublicKeySize*2, 10) + `) NOT NULL,
                    AuthToken TEXT NOT NULL,
                    LastNotifySend BIGINT UNSIGNED NOT NULL DEFAULT 0,
                    LastNotifyFrom BIGINT UNSIGNED NOT NULL DEFAULT 0,
                    LastFetch BIGINT UNSIGNED NOT NULL DEFAULT 0,
                    ErrorCount BIGINT UNSIGNED NOT NULL DEFAULT 0,
                    LastPosition BIGINT UNSIGNED NOT NULL DEFAULT 0,
                    UNIQUE KEY pk(PublicKey)
                );`,
			"InsertPeer":       `INSERT INTO peer (PublicKey) VALUES (?);`,
			"UpdateStatPeer":   `UPDATE peer SET LastFetch=?, LastPosition=?, ErrorCount=? WHERE PublicKey=?;`,
			"UpdateNotifyPeer": `UPDATE peer SET LastNotifySend=?, ErrorCount=ErrorCount+? WHERE PublicKey=?;`,
			"UpdateTokenPeer":  `UPDATE peer SET LastNotifyFrom=?, Authtoken=? WHERE PublicKey=?;`,
			"SelectPeer":       `SELECT AuthToken, LastNotifySend, LastNotifyFrom, LastFetch, ErrorCount, LastPosition FROM peer WHERE PublicKey=?;`,
			"MessageCreate": `CREATE TABLE IF NOT EXISTS message (
                    ID BIGINT UNSIGNED NOT NULL PRIMARY KEY AUTO_INCREMENT,
                    Counter BIGINT UNSIGNED NOT NULL DEFAULT 0,
                    MessageID VARCHAR(` + strconv.FormatInt(message.MessageIDSize*2, 10) + `) NOT NULL,
                    ReceiverConstantPubKey VARCHAR(` + strconv.FormatInt(message.Curve25519KeySize*2, 10) + `) NOT NULL,
                    SignerPub VARCHAR(` + strconv.FormatInt(message.SignerPubKeySize*2, 10) + `) NOT NULL,
                    PostTime BIGINT UNSIGNED NOT NULL DEFAULT 0,
                    ExpireTime BIGINT UNSIGNED NOT NULL DEFAULT 0,
                    ExpireRequest BIGINT UNSIGNED NOT NULL DEFAULT 0,
                    Distance BIGINT UNSIGNED NOT NULL DEFAULT 0,
                    OneTime TINYINT UNSIGNED NOT NULL DEFAULT 1,
                    Sync TINYINT UNSIGNED NOT NULL DEFAULT 0,
                    Hidden TINYINT UNSIGNED NOT NULL DEFAULT 1,
                    UNIQUE INDEX keyCount (Counter, ReceiverConstantPubKey),
                    UNIQUE KEY mid (MessageID)
                );`,
			"InsertMessage": `INSERT INTO message
                    (Counter, MessageID, ReceiverConstantPubKey, SignerPub,
                    PostTime, ExpireTime, ExpireRequest, Distance, OneTime, Sync, Hidden)
                    VALUES ( ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
                ;`,
			"SelectMessage": `SELECT ID, Counter, MessageID, ReceiverConstantPubKey, SignerPub,
                    PostTime, ExpireTime, ExpireRequest, Distance, OneTime, Sync, Hidden FROM message
                    WHERE MessageID=?;`,
			"DeleteMessage":       `DELETE FROM message WHERE MessageID=?;`,
			"UpdateExpireMessage": `UPDATE message SET ExpireTime=? WHERE MessageID=?;`,
			"SelectExpireMessage": `SELECT MessageID, SignerPub FROM message WHERE ExpireTime<?;`,
			"MessageCounterCreate": `CREATE TABLE IF NOT EXISTS messageCounter (
                ReceiverConstantPubKey VARCHAR(` + strconv.FormatInt(message.Curve25519KeySize*2, 10) + `) NOT NULL,
                Counter BIGINT UNSIGNED NOT NULL DEFAULT 1,
                LastTime BIGINT UNSIGNED NOT NULL DEFAULT 0,
                UNIQUE INDEX Receiver(ReceiverConstantPubKey),
                KEY c (Counter)
                );`,
			"NextMessageCounter":     `SELECT MAX(Counter) FROM messageCounter WHERE ReceiverConstantPubKey=?;`,
			"IncreaseMessageCounter": `UPDATE messageCounter SET Counter=Counter+1, LastTime=? WHERE ReceiverConstantPubKey=?;`,
			"InsertMessageCounter":   `INSERT INTO messageCounter (ReceiverConstantPubKey, LastTime) VALUES (?,? );`,
			"ExpireMessageCounter":   `DELETE FROM messageCounter WHERE LastTime<?;`,
			"GlobalIndexCreate": `CREATE TABLE IF NOT EXISTS globalindex (
                    ID BIGINT UNSIGNED NOT NULL PRIMARY KEY AUTO_INCREMENT,
                    Message BIGINT UNSIGNED NOT NULL,
                    EntryTime BIGINT UNSIGNED NOT NULL DEFAULT 0,
                    UNIQUE KEY Message(Message),
                    FOREIGN KEY (Message) REFERENCES message(ID) ON DELETE CASCADE
                );`,
			"globalIndexAdd": `INSERT INTO globalindex (Message, EntryTime) VALUES (?, ?);`,
			"getKeyIndex": `SELECT ID, Counter, MessageID, ReceiverConstantPubKey, SignerPub,
                    PostTime, ExpireTime, ExpireRequest, Distance, OneTime, Sync, Hidden FROM message
                    WHERE ReceiverConstantPubKey=? AND Counter>? ORDER BY Counter ASC LIMIT ?
                ;`,
			"getGlobalIndex": `SELECT m.ID, i.ID, m.MessageID, m.ReceiverConstantPubKey, m.SignerPub,
                    m.PostTime, m.ExpireTime, m.ExpireRequest, m.Distance, m.OneTime, m.Sync, m.Hidden 
                    FROM message AS m, globalindex AS i
                    WHERE i.ID>? ORDER BY i.ID ASC LIMIT ?
                ;`,
			"messageBlobCreate": `CREATE TABLE IF NOT EXISTS messageblob (
                    Message BIGINT UNSIGNED NOT NULL,
                    MessageID VARCHAR(` + strconv.FormatInt(message.MessageIDSize*2, 10) + `) NOT NULL,
                    SignerPub VARCHAR(` + strconv.FormatInt(message.SignerPubKeySize*2, 10) + `) NOT NULL,
                    OneTime INT NOT NULL DEFAULT 0,  
                    DATA MEDIUMBLOB,
                    UNIQUE KEY Message(Message),
                    FOREIGN KEY (Message) REFERENCES message(ID) ON DELETE CASCADE
                );`,
			"messageBlobInsert": `INSERT INTO messageblob 
                    (Message, MessageID, SignerPub, OneTime, Data) VALUES
                    (?, ?, ?, ?, ?);`,
			"messageBlobSelect": `SELECT Message, MessageID, SignerPub, OneTime, Data FROM messageblob WHERE MessageID=?;`,
			"messageBlobDelete": `DELETE FROM messageblob WHERE MessageID=?;`,
		},
		"sqlite3": map[string]string{
			"SignerCreate": `CREATE TABLE IF NOT EXISTS signer (
                    ID INTEGER PRIMARY KEY,
                    PublicKey VARCHAR(` + strconv.FormatInt(message.SignerPubKeySize*2, 10) + `) NOT NULL,
                    Nonce VARCHAR(` + strconv.FormatInt(hashcash.NonceSize*2, 10) + `) NOT NULL,
                    Bits INT NOT NULL DEFAULT 0,
                    MessagesPosted BIGINT NOT NULL DEFAULT 0,
                    MessagesRetained BIGINT NOT NULL DEFAULT 0,
                    MaxMessagesPosted BIGINT UNSIGNED NOT NULL DEFAULT 0,
                    MaxMessagesRetained BIGINT UNSIGNED NOT NULL DEFAULT 0,
                    ExpireTarget BIGINT UNSIGNED NOT NULL DEFAULT 0,
                    LastMessageDeleted BIGINT UNSIGNED NOT NULL DEFAULT 0,
                    UNIQUE (PublicKey)
                );`,
			"SignerInsert": `INSERT INTO signer (PublicKey,Nonce,Bits,
                    MaxMessagesPosted,MaxMessagesRetained,ExpireTarget) VALUES
                    (?,?,?,?,?,?)
                ;`,
			"SelectSignerPublicKey": `SELECT ID, PublicKey, Nonce, Bits, MessagesPosted,
                    MessagesRetained, MaxMessagesPosted, MaxMessagesRetained, ExpireTarget FROM
                    signer WHERE PublicKey=?
                ;`,
			"SelectSignerID": `SELECT ID, PublicKey, Nonce, Bits, MessagesPosted,
                    MessagesRetained, MaxMessagesPosted, MaxMessagesRetained, ExpireTarget FROM
                    signer WHERE ID=?
                ;`,
			"UpdateSigner": `UPDATE signer SET Nonce=?, Bits=?, MaxMessagesPosted=?, 
                    MaxMessagesRetained=?, ExpireTarget=? WHERE PublicKey=?
                ;`,
			// "UpdateOrInsertSigner": ``,
			"AddMessageSigner": `UPDATE signer 
                SET MessagesPosted=MessagesPosted+1, MessagesRetained=MessagesRetained+1
                WHERE PublicKey=? AND MaxMessagesRetained>MessagesRetained AND MaxMessagesPosted>MessagesPosted
                ;`,
			"DelMessageSigner": `UPDATE signer
                SET MessagesRetained=MessagesRetained-1
                WHERE PublicKey=? 
                ;`,
			"PrepareExpireSigner": `UPDATE signer SET LastMessageDeleted=? WHERE 
                MessagesRetained<=0 AND LastMessageDeleted=0
                ;`,
			"DeleteExpireSigner": `DELETE FROM signer 
                WHERE LastMessageDeleted!=0 AND LastMessageDeleted<?
                ;`,
			"PeerCreate": `CREATE TABLE IF NOT EXISTS peer (
                    ID INTEGER PRIMARY KEY,
                    PublicKey VARCHAR(` + strconv.FormatInt(ed25519.PublicKeySize*2, 10) + `) NOT NULL,
                    AuthToken TEXT NOT NULL DEFAULT "",
                    LastNotifySend BIGINT UNSIGNED NOT NULL DEFAULT 0,
                    LastNotifyFrom BIGINT UNSIGNED NOT NULL DEFAULT 0,
                    LastFetch BIGINT UNSIGNED NOT NULL DEFAULT 0,
                    ErrorCount BIGINT UNSIGNED NOT NULL DEFAULT 0,
                    LastPosition BIGINT UNSIGNED NOT NULL DEFAULT 0,
                    UNIQUE (PublicKey)
                );`,
			"InsertPeer":       `INSERT INTO peer (PublicKey) VALUES (?);`,
			"UpdateStatPeer":   `UPDATE peer SET LastFetch=?, LastPosition=?, ErrorCount=? WHERE PublicKey=?;`,
			"UpdateNotifyPeer": `UPDATE peer SET LastNotifySend=?, ErrorCount=ErrorCount+? WHERE PublicKey=?;`,
			"UpdateTokenPeer":  `UPDATE peer SET LastNotifyFrom=?, Authtoken=? WHERE PublicKey=?;`,
			"SelectPeer":       `SELECT AuthToken, LastNotifySend, LastNotifyFrom, LastFetch, ErrorCount, LastPosition FROM peer WHERE PublicKey=?;`,
			"MessageCreate": `CREATE TABLE IF NOT EXISTS message (
                    ID INTEGER PRIMARY KEY,
                    Counter BIGINT UNSIGNED NOT NULL DEFAULT 0,
                    MessageID VARCHAR(` + strconv.FormatInt(message.MessageIDSize*2, 10) + `) NOT NULL,
                    ReceiverConstantPubKey VARCHAR(` + strconv.FormatInt(message.Curve25519KeySize*2, 10) + `) NOT NULL,
                    SignerPub VARCHAR(` + strconv.FormatInt(message.SignerPubKeySize*2, 10) + `) NOT NULL,
                    PostTime BIGINT UNSIGNED NOT NULL DEFAULT 0,
                    ExpireTime BIGINT UNSIGNED NOT NULL DEFAULT 0,
                    ExpireRequest BIGINT UNSIGNED NOT NULL DEFAULT 0,
                    Distance BIGINT UNSIGNED NOT NULL DEFAULT 0,
                    OneTime TINYINT UNSIGNED NOT NULL DEFAULT 1,
                    Sync TINYINT UNSIGNED NOT NULL DEFAULT 0,
                    Hidden TINYINT UNSIGNED NOT NULL DEFAULT 1,
                    UNIQUE (Counter, ReceiverConstantPubKey),
                    UNIQUE (MessageID)
                );`,
			"InsertMessage": `INSERT INTO message
                    (Counter, MessageID, ReceiverConstantPubKey, SignerPub,
                    PostTime, ExpireTime, ExpireRequest, Distance, OneTime, Sync, Hidden)
                    VALUES ( ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
                ;`,
			"SelectMessage": `SELECT ID, Counter, MessageID, ReceiverConstantPubKey, SignerPub,
                    PostTime, ExpireTime, ExpireRequest, Distance, OneTime, Sync, Hidden FROM message
                    WHERE MessageID=?;`,
			"DeleteMessage":       `DELETE FROM message WHERE MessageID=?;`,
			"UpdateExpireMessage": `UPDATE message SET ExpireTime=? WHERE MessageID=?;`,
			"SelectExpireMessage": `SELECT MessageID, SignerPub FROM message WHERE ExpireTime<?;`,
			"MessageCounterCreate": `CREATE TABLE IF NOT EXISTS messageCounter (
                ReceiverConstantPubKey VARCHAR(` + strconv.FormatInt(message.Curve25519KeySize*2, 10) + `) NOT NULL,
                Counter BIGINT UNSIGNED NOT NULL DEFAULT 1,
                LastTime BIGINT UNSIGNED NOT NULL DEFAULT 0,
                UNIQUE (ReceiverConstantPubKey),
                UNIQUE (ReceiverConstantPubKey, Counter)
                );`,
			"NextMessageCounter":     `SELECT MAX(Counter) FROM messageCounter WHERE ReceiverConstantPubKey=?;`,
			"IncreaseMessageCounter": `UPDATE messageCounter SET Counter=Counter+1, LastTime=? WHERE ReceiverConstantPubKey=?;`,
			"InsertMessageCounter":   `INSERT INTO messageCounter (ReceiverConstantPubKey, LastTime) VALUES (?,? );`,
			"ExpireMessageCounter":   `DELETE FROM messageCounter WHERE LastTime<?;`,
			"GlobalIndexCreate": `CREATE TABLE IF NOT EXISTS globalindex (
                    ID INTEGER PRIMARY KEY,
                    Message BIGINT UNSIGNED NOT NULL,
                    EntryTime BIGINT UNSIGNED NOT NULL DEFAULT 0,
                    UNIQUE (Message),
                    FOREIGN KEY (Message) REFERENCES message(ID) ON DELETE CASCADE
                );`,
			"globalIndexAdd": `INSERT INTO globalindex (Message, EntryTime) VALUES (?, ?);`,
			"getKeyIndex": `SELECT ID, Counter, MessageID, ReceiverConstantPubKey, SignerPub,
                    PostTime, ExpireTime, ExpireRequest, Distance, OneTime, Sync, Hidden FROM message
                    WHERE ReceiverConstantPubKey=? AND Counter>? ORDER BY Counter ASC LIMIT ?
                ;`,
			"getGlobalIndex": `SELECT m.ID, i.ID, m.MessageID, m.ReceiverConstantPubKey, m.SignerPub,
                    m.PostTime, m.ExpireTime, m.ExpireRequest, m.Distance, m.OneTime, m.Sync, m.Hidden 
                    FROM message AS m, globalindex AS i
                    WHERE i.ID>? ORDER BY i.ID ASC LIMIT ?
                ;`,
			"messageBlobCreate": `CREATE TABLE IF NOT EXISTS messageblob (
                    Message BIGINT UNSIGNED NOT NULL,
                    MessageID VARCHAR(` + strconv.FormatInt(message.MessageIDSize*2, 10) + `) NOT NULL,
                    SignerPub VARCHAR(` + strconv.FormatInt(message.SignerPubKeySize*2, 10) + `) NOT NULL,
                    OneTime INT NOT NULL DEFAULT 0,  
                    DATA BLOB,
                    UNIQUE (Message),
                    FOREIGN KEY (Message) REFERENCES message(ID) ON DELETE CASCADE
                );`,
			"messageBlobInsert": `INSERT INTO messageblob 
                    (Message, MessageID, SignerPub, OneTime, Data) VALUES
                    (?, ?, ?, ?, ?);`,
			"messageBlobSelect": `SELECT Message, MessageID, SignerPub, OneTime, Data FROM messageblob WHERE MessageID=?;`,
			"messageBlobDelete": `DELETE FROM messageblob WHERE MessageID=?;`,
		},
	}
)
