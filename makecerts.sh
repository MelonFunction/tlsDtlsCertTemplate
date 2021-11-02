cd certs

# Extensions required for certificate validation.
EXTFILE='extfile.conf'
echo 'subjectAltName = IP:127.0.0.1\nbasicConstraints = critical,CA:true' > "${EXTFILE}"
# Server.
SERVER_NAME='server'
openssl ecparam -name prime256v1 -genkey -noout -out "${SERVER_NAME}.pem"
openssl req -key "${SERVER_NAME}.pem" -new -sha256 -subj '/C=NL' -out "${SERVER_NAME}.csr"
openssl x509 -req -in "${SERVER_NAME}.csr" -extfile "${EXTFILE}" -days 365 -signkey "${SERVER_NAME}.pem" -sha256 -out "${SERVER_NAME}.pub.pem"