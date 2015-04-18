package main

/*

This program implements the second Go Challenge: establishing secure communication
between a server and a client, using asymmetric encryption.


Implementation details


When a connection is established, both ends (the client and the server) receive
a net.Conn structure and NewEncryptedConnection is called. This function generates
a new key pair and attempts to perform the key exchange (send your own public key
and read the other's).

If it was successful, an EncryptedConnection structure (which implements io.ReadWriteCloser)
is created, delegating reads and writes to a SecureReader (implements io.Reader,
performs on the fly decryption) and SecureWriter (implements io.Writer, performs
on the fly encryption).

The encryption function require a nonce (unique, unrepeatable "number") to be used. In
order to keep the code simpler and reduce the on-wire overhead, SecureWriter generates
a random nonce on the first Write(); this won't be regenerated during the lifetime of
the writer.

SecureReader has a similar behaviour, on the first Read() it attempts to read the
nonce and then the encrypted data.


The communication flow looks like this:



    Client                              Server

      |                                    |
      |                               Listen()
      |                                    |
Dial("server") --------------------------> |
      |                               Accept()
      |                                    |
connected                             connected
      |                                    |
NewEncryptedConnection()              NewEncryptedConnection()
  cPub, cPriv := gen key pair            sPub, sPriv := gen key pair
      |                                    |
  Write(cPub) ----------\ /------------  Write(sPub)
      |                  X                 |
  Read(sPub)  <---------/ \------------> Read(cPub)
      |                                    |
      |                                    |
Write("some message")                 Read()
      |                                    |
  generate nonce                           |
  write nonce ------------------------> read nonce
      |                                    |
  encrypt message                          |
  write encryptedMessage -------------> read encrypted message
      |                                 decrypt
  close


*/
