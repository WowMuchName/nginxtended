if [ $# -ne 2 ]; then echo "Usage make-user-cert.sh {User} {CertDirectory}"; exit 1; fi

certFile=$2"/cert.pem"
userid=$(echo "$1" | sed -n "s/ //p")
serial=$(head -200 /dev/urandom | cksum  | cut -f1 -d " ")

if [ ! -f "$certFile" ]; then
  echo $certFile" does not exist"
  exit 1
fi

domain=$(openssl x509 -in "$certFile" -subject | sed -n "s/^.*CN=\(.*\)$/\1/p")
clientAuthCertFile=$2"/"$domain".client.crt"
clientAuthKeyFile=$2"/"$domain".client.key"

if [ ! -f "$clientAuthCertFile" ]; then
  echo $clientAuthCertFile" does not exist"
  exit 1
fi

if [ ! -f "$clientAuthKeyFile" ]; then
  echo $clientAuthKeyFile" does not exist"
  exit 1
fi

echo "Creating client-authentication for <"$userid"@"$domain">"

# Read Password
echo -n "Password: "
read -s password
echo
echo -n "Password (repeat): "
read -s passwordRepeat
echo

if [ "$password" != "$passwordRepeat" ]; then
  echo "Password missmatch"
  exit 1
fi

openssl req -out $userid".csr" -new -newkey rsa:2048 -nodes -keyout $userid".key" -subj "/O=$domain/CN=$1/" -passout pass:"$password"
openssl x509 -req -days 365 -in $userid".csr" -CA "$clientAuthCertFile" -CAkey "$clientAuthKeyFile" -set_serial 01 -out $userid".crt" -passin pass:"$password"
openssl pkcs12 -export -out $userid".p12" -inkey $userid".key" -in $userid".crt" -certfile "$clientAuthCertFile" -passout pass:"$password"
rm $userid".csr" $userid".crt" $userid".key"
