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
		dht->setup(pins[i], DHT::DHT22);
		pinfo->data.pointers[1] = dht;
	}
}

static int fix_temp(float value) {
	return round((value + 50) * 10);
}

static void dht_on_measure(PSensorInfo pinfo, POutputBuffer buffer) {
	DHT *dht = (DHT *)pinfo->data.pointers[1];
	float hum = dht->getHumidity();
	float temp = dht->getTemperature();
	root_add_measure(buffer, pinfo, 0, fix_temp(hum));
	root_add_measure(buffer, pinfo, 1, fix_temp(temp));
}
