The following credential files need to be generated within this directory for
the mosquitto container to work:

./mosquitto.passwd
Username and password file for moquitto authentication. Generate with your
desired crednetials using the mosquitto-passwd command.

./pki/ca.crt
TLS cert for the CA that signed your broker's cert

./pki/broker.crt
TLS cert to use for the broker

./pki/broker.key
Private key associated with the broker's TLS cert