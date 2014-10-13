#include "lib_dht.h"

#include "AutoHome.h"
#include <math.h>

static void dht_on_create(PSensorTypeInfo info) {
	byte pins[DHT_PINS_COUNT] = DHT_PINS;
	int i;
	for (i = 0; i < DHT_PINS_COUNT; i++) {
		// Create
		PSensorInfo pinfo = root_new_sensor_info(info, i);
		pinfo->data.bytes[0] = pins[i];
		DHT *dht = new DHT();
		dht->setup(pins[i], DHT::DHT11);
		pinfo->data.pointers[1] = dht;
	}
}

static void dht_on_measure(PSensorInfo pinfo, POutputBuffer buffer) {
	DHT *dht = (DHT *)pinfo->data.pointers[1];
	byte hum = (byte)round(dht->getHumidity());
	byte temp = (byte)round(dht->getTemperature());
	if (hum>0) {
		root_add_measure(buffer, pinfo, 0, hum);
	}
	if (temp != 0) {
		root_add_measure(buffer, pinfo, 1, temp);
	}
}
