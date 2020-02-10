# Twofer

A service implementing some two factor authentication stuff
 
 
 
 ## Freja
 * Get a .pfx file and password for it
 * Decrypt it `openssl pkcs12 -in mf.pfx -out all.pem -nodes`
 * Create tree file, rootca.pem, cert.pem, key.pem
 * From all.pem copy portion "Integration Test PRIVATE KEY" into key.pem
 * From all.pem copy portion "Integration Test CERTIFICATE" into cert.pem
 * From all.pem copy portion "RSA Test Root CA CERTIFICATE" and "RSA TEST Issuing CA CERTIFICATE" into rootca.pem



## Documentation suggestion
 * For better explanation of how a client mTLS should work..
 * Add explanation if a content type is needed
 
 