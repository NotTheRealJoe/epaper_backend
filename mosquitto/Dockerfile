FROM eclipse-mosquitto

COPY mosquitto.conf /mosquitto/config/mosquitto.conf
COPY mosquitto.passwd /mosquitto/mosquitto.passwd

COPY pki/ca.crt /mosquitto/tls/
# will need to be updated when host changes
COPY pki/broker.crt /mosquitto/tls/broker.crt
COPY pki/broker.key /mosquitto/tls/broker.key