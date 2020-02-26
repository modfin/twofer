# Twofer
A stateless service implementing some two factor authentication methods, so life is gets easier. 
 
 ## General 
 Twofer is intended to be deployed within your stack and not be accessible directly from the outside. 
 Instead Twofer can be configured to expose gRPC APIs in order to handle different factors in your multi part authentication 
 scheme.
 
 
## E-ID
Twofer support Swedish BankId and FrejaID as Electronic identification. This can be used for signup in order to collect the 
 identity of the user, as a factor in a authentication scheme or for collecting signatures.
 
### API 
The gRPC API for Authentication and Signatures abstracts the provider that is being used and unifies the way of handling 
E-ID.  

There is 5 method calls
* `GetProviders` - Returns a list of active Eid providers registered. eg. BankId and/or FrejaId
* `AuthInit` - Initiates a Authentication request  
* `SignInit` - Initiates a Signature request
* `Peek` - Returns the current status of a Auth or a Sign request
* `Collect` - Waits for a Auth or a Sign request to finish and returns the result
* `Cancel` - Cancels an ongoing request 

### Swedish BankID - [bankid.com](https://www.bankid.com/bankid-i-dina-tjanster/rp-info)
Twofer is in the context of BankID considered a Relying party.

#### For testing
* Download the a certificate SSL certificate for test [FPTestcert2_20150818_102329.pfx](https://www.bankid.com/assets/bankid/rp/FPTestcert2_20150818_102329.pfx)
* Extract the pem file `openssl pkcs12 -in FPTestcert2_20150818_102329.pfx -out bank_id_all.pem -nodes` (password: qwerty123)
* From bank_id_all.pem, extract Private Key portion into  `bank-id-key.pem`
* From bank_id_all.pem, extract Certificate portion into  `bank-id-cert.pem` 
* From [documentation](https://www.bankid.com/assets/bankid/rp/bankid-relying-party-guidelines-v3.2.2.pdf) copy Root CA pem (section 8) into `bank-id-rootca.pem`

**Config**
When starting twofer add the following environment variables
```bash
EID_BANKID_ENABLE=true
EID_BANKID_URL=https://appapi2.test.bankid.com
EID_BANKID_ROOT_CA_PEM_FILE=/path/to/bank-id-rootca.pem  ## Used to authenticate BankID servers servers towards twofer
EID_BANKID_CLIENT_CERT_FILE=/path/to/bank-id-cert.pem    ## Used to authenticate your account towards BankID
EID_BANKID_CLIENT_KEY_FILE=/path/to/bank-id-key.pem      ## Used to authenticate your account towards BankID
```

**Use**
* Go to https://demo.bankid.com/ and register a test account.
* Use gRPC client.


### FrejaID - [frejaeid.com](https://org.frejaeid.com/en/developers-section)
Twofer is in the context of FrejaID considered a Relying party.

#### For testing
* Request a .pfx file from Frejas customer service
* Extract the pem file `openssl pkcs12 -in freja.pfx -out freja_all.pem -nodes` (password: qwerty123)
* From freja_all.pem, extract Private Key portion into  `freja-key.pem`
* From freja_all.pem, extract Certificate portion into  `freja-cert.pem` 
* From [documentation](https://frejaeid.com/rest-api/Freja%20eID%20Relying%20Party%20Developers'%20Documentation.html#FrejaeIDRelyingPartyDevelopers'Documentation-ServerSSLcertificate) 
copy Root CA pem into `freja-rootca.pem`
* From [documentation](https://frejaeid.com/rest-api/Freja%20eID%20Relying%20Party%20Developers'%20Documentation.html#FrejaeIDRelyingPartyDevelopers'Documentation-JWSJWScertificate) 
copy JWS pem into `freja-jws-cert.pem`


**Config**
When starting twofer add the following environment variables
```bash
EID_FREJA_ENABLE=true
EID_FREJA_URL=https://services.test.frejaeid.com
EID_FREJA_ROOT_CA_PEM_FILE=/path/to/freja-rootca.pem   ## Used to authenticate Freja servers towards twofer
EID_FREJA_CLIENT_CERT_FILE=/path/to/freja-cert.pem     ## Used to authenticate your account towards Freja
EID_FREJA_CLIENT_KEY_FILE=/path/to/freja-key.pem       ## Used to authenticate your account towards Freja
EID_FREJA_JWS_CERT_FILE=/path/to/freja-jws-cert.pem    ## Used to verify messages sent by Freja
```

 
## OTP
TOTP and HOTP is often part of a multi factor scheme and while this is often not hard to implement, it might be harder 
to protect and there are a few consideration when implementing it. There for twofer includes a OTP service that helps 
with enrollment and verification

**Config**
* Generate a AES key, eg `$ echo 1:aes:$(openssl rand -base64 16)`

```bash
OTP_ENABLE="true"
OTP_ENCRYPTION_KEY="1:aes:Hg44JefQsFJMI1F0zhWMpw=="  # Used to seal and open the uri in order not to stor it in plain text
```


**Use**
* gRPC `Enroll`, persist returning secret in a database coupled with the user
* gRPC `Validate`, update/persist returning secret in database coupled with the user

 
## WebAuthn
 
## QR
Since both BankID, Freja and OTP have QR-code components, a gRPC api is included which turns test in to a png QR code image

**Use**
* gRPC `Generate`


 