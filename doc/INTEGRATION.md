## Integrating repbin into your own scripts

Repclient includes some aids to help with application integration. All
communication between caller and repclient happens through the repclient
parameters and file descriptors.

To get started, enable the "--verbose" and "--appdata" parameters when calling
repclient. This will enabled output on  stderr that is suitable for caller
interaction. The format is "STATUS (subject):\tDATA\n". The subject of the
message is always encapsulated in brackets () and the data always follows a tab.

Input and output of file/post data should happen through the use of stdin/stdout
pipes, alternatively the --in and --out parameters accept a number which is
interpreted as a file descriptor. File descriptors will be read/written ONLY
when needed and exactly WHILE needed. As soon as repclient knows that it will
not further use the file descriptor it will close it.


## Key management

Repclient does not implement any key management itself. Instead it makes keys
available to calling applications (see below) and requests keys from the calling
application.

To request keys, two mutually exclusive methods are vailable. The use of
--privkey=FileDescriptor (where FileDescriptor is a number) allows only the
provision of keys before the message is parsed. In this case the calling
operation must already know which private key matches the message. Repclient
will send "STATUS(KeyMGT): ENTER KEY" over stderr before starting to read the
key.

The second method is interactive key requests from the caller. It requires
--keymgt=FileDescriptor as parameter when calling repclient. For each public key
for which a private key is required, repclient will send "STATUS(KeyMGTRequest):
$PublicKey$" on stderr. It then starts reading from the file descriptor.

Success/Erro is signalled by:
```
	STATUS(KeyMGT): READ DONE
	STATUS(KeyMGT): READ FAILURE
```

The --privkey=FileDescriptor method should be used for encryption operations. --keymgt=FileDescriptor is ONLY available for the decrypt operation.

If repclient creates keys, the following status output is available:
Embedded keys (auto-generated):
```
	STATUS(EmbedPrivateKey): $ConstantPrivKey$_$TemporaryPrivKey$
	STATUS(EmbedPublicKey): $ConstantPubKey$_$TemporaryPubKet$
```

Key generation, posting:
```
	STATUS(PrivateKey): $PrivateKey$
	STATUS(PublicKey): $PublicKey$
	STATUS(PrivateKey): $ConstantPrivateKey$_$TemporaryPrivateKey$
	STATUS(PublicKey): $ConstantPublicKey$_$TemporaryPublicKey$
```

## Some application data output:

Fetching messages:
```
	STATUS(FetchComplete): $MessageID$
	STATUS(FetchError): $MessageID$
	STATUS(FetchResult): ERROS $count$
	STATUS(FetchResult): OK 
	STATUS(FetchServer): $Server$
	STATUS(Fetch): $MessageID$
```

List messages (message-type 2):
```
	STATUS(ListInput): NULL $MessageID$ NULL
	STATUS(ListInput): NULL $MessageID$ $PrivKey$
```

Post-Box getindex:
```
	STATUS(ListResult): $Start$ $Count$ $MoreMessages$
	STATUS(MessageList): $Counter$ $MessageID$ $SignerPubKey$ $PostTime$ $ExpireTime$
```

Posting messages:
```
	STATUS(Message): $MessageID$_$PrivKey$
	STATUS(URL): $Server$/$MessageID$_$PrivateKey$
	STATUS(RecPubKey): $ConstantPublicKey$
```

Fetching messages:
```
	STATUS(PubKeySig): $PublicKey$
	STATUS(SenderPubKey): $ConstantPublicKey$
	STATUS(RecPubKey): $ConstantPublicKey$
```

Repost message handling:
```
	STATUS(RepostSettings): $MinDelay$ $MaxDelay$
	STATUS(STMCleanErr): $Error$
	STATUS(STMErr): $File$ $Error$
	STATUS(STMFile): $File$
	STATUS(STMKeep): $File$
	STATUS(STMRes): $File$ DONE
	STATUS(STMRes): $File$ FAIL
	STATUS(STMSend): $File$
	STATUS(STM): $MinDelay$ $MaxDelay$ $SendTime$
	STATUS(STMTrans): $File$
```

