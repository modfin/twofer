# Twofer
A stateless service implementing some two factor authentication methods, so life is gets somewhat easier. 
 
 ## General 
 Twofer is intended to be deployed within your stack and not be accessible directly from the outside. 
 Instead Twofer can be configured to expose gRPC APIs in order to handle different factors in your multi part authentication 
 scheme.
 
 
## E-ID
Twofer support Swedish BankId as Electronic identification. This can be used for signup in order to collect the 
 identity of the user, as a factor in a authentication scheme or for collecting signatures.
 
### API 
The gRPC API for Authentication and Signatures abstracts the provider that is being used and unifies the way of handling 
E-ID.  

There is 5 method calls
* `GetProviders` - Returns a list of active Eid providers registered. eg. BankId
* `AuthInit` - Initiates a Authentication request  
* `SignInit` - Initiates a Signature request
* `Peek` - Returns the current status of a Auth or a Sign request
* `Collect` - Waits for a Auth or a Sign request to finish and returns the result
* `Cancel` - Cancels an ongoing request 

### Swedish BankID - [bankid.com](https://www.bankid.com/bankid-i-dina-tjanster/rp-info)
Twofer is in the context of BankID considered a Relying party.

#### For testing
* Download the a certificate SSL certificate for test [FPTestcert3_20200618.p12](https://www.bankid.com/assets/bankid/rp/FPTestcert3_20200618.p12)
* Extract the pem file `openssl pkcs12 -in FPTestcert3_20200618.p12 -out bank_id_all.pem -nodes` (password: qwerty123)
* From bank_id_all.pem, extract Private Key portion into  `bank-id-key.pem`
* From bank_id_all.pem, extract Certificate portion into  `bank-id-cert.pem` 
* From [documentation](https://www.bankid.com/assets/bankid/rp/bankid-relying-party-guidelines-v3.5.pdf) copy Root CA pem (section 8) into `bank-id-rootca.pem`

**Config**
When starting twofer add the following environment variables
```bash
EID_BANKID_ENABLE=true
EID_BANKID_URL=https://appapi2.test.bankid.com

## Used to authenticate BankID servers servers towards twofer
## EID_BANKID_ROOT_CA_PEM can be used to load pm directly from file
EID_BANKID_ROOT_CA_PEM_FILE=/path/to/bank-id-rootca.pem  

## Used to authenticate your account towards BankID
## EID_BANKID_CLIENT_CERT can be used to load pm directly from file
EID_BANKID_CLIENT_CERT_FILE=/path/to/bank-id-cert.pem    

## Used to authenticate your account towards BankID
## EID_BANKID_CLIENT_KEY can be used to load pm directly from file
EID_BANKID_CLIENT_KEY_FILE=/path/to/bank-id-key.pem      
```

**Use**
* Go to https://demo.bankid.com/ and register a test account.
* Use gRPC client.

## OTP
TOTP and HOTP is often part of a multi factor scheme and while this is often not hard to implement, it might be harder 
to protect and there are a few consideration when implementing it. There for twofer includes a OTP service that helps 
with enrollment and authentication

**State**
The OTP relies on a state in order to verify a user, this means the user must persist the userblob when provided. This 
since the state must be passed to twofer when called

**Config**
* Generate a AES key, eg `$ echo 1:aes:$(openssl rand -base64 16)`

```bash
OTP_ENABLE="true"

# Used to seal and open the uri in order not to stor it in plain text
# The latest key version is always used, this means that on each Auth the returning blob
# will be encrypted using this key, this enables you to upgrade keys that protects the users credentials
OTP_ENCRYPTION_KEY="1:aes:Hg44JefQsFJMI1F0zhWMpw== 2:aes:xzK4KyrOUo45VfFiF9vijw=="  

# How many attempts verification attempts can be made for the same user a minute
OTP_RATE_LIMIT=10 # Default: 10 `

# If counter mode is used, How many OTPs ahead is checked
OTP_SKEW_COUNTER=5 # Default: 5 `

# If time mode is used, How many OTPs forward and backwards in time is checked
OTP_SKEW_TIME=1 # Default: 1`
```

**Use**
* gRPC `Enroll`, persist returning userBlob in a database coupled with the user
* gRPC `Auth`, update/persist returning userBlob in database coupled with the user

 
## WebAuthn
WebAuthn is a protocol to verify a user through, among other thins, the browser by eg. using a FIDO2 key.
See https://webauthn.io/ or https://www.w3.org/TR/webauthn/ for more information. While 

**State**
The WebAuthn relies on a state in order to verify a user, this means the user must persist the userblob when provided. This 
since the state must be passed to twofer when called

**Config**
* Generate a HMAC key, eg `$ echo $(openssl rand -base64 32)`

```bash
WEBAUTHN_ENABLED=true
WEBAUTHN_RP_ID=localhost
WEBAUTHN_RP_DISPLAYNAME=localhost
WEBAUTHN_RP_ORIGIN=http://localhost:8080
WEBAUTHN_HMAC_KEY=+SoWOS6kLTe8OOVTBXnQ+lMAsUH0hncsnCJUQ2javqw=

# Can be discouraged/proffered/required 
WEBAUTHN_USER_VERIFICATION=discouraged # Default: discouraged `

# How many api calls can be made for the same user per minute
WEBAUTHN_RATE_LIMIT=10 # Default: 10

# Once a session is issues, for how long is it valid
WEBAUTHN_TIMEOUT=60s # Default: 60s
```

**Use**
* gRPC `EnrollInit`, pass in the current userBlob, if it exist. It creates a session and json, the json shall be passed to the frontend for the authenticator to interact with. 
* gRPC `EnrollFinal`, the session that from EnrollInit creates shall be passed coupled with the signature (the frontend response). 
  This returns a userBlob on success, this blob should be persisted and used in the auth requests. If a user blob existed prior to the enrollment it 
  shall be replaced by the returning one. This allows for a user to have multiple authenticators.
* gRPC `AuthInit`, pass in the current userBlob. It creates a session and json, the json shall be passed to the frontend for the authenticator to interact with.
* gRPC `AuthFinal`, the session that from AuthInit creates shall be passed coupled with the signature (the frontend response). 
  If successful it returns valid = true
}


## QR
Since both BankID and OTP have a QR-code components, a gRPC api is included which turns test in to a png QR code image

**Use**
* gRPC `Generate`

## Dev
Release new versions of the Docker image onto [Dockerhub](https://hub.docker.com/r/modfin/twofer)  
```
docker login --username=yourhubusername # then enter pass
./docker-build-push.sh
```
