#include "AutoHome.h"

static void movement_on_create(PSensorTypeInfo info) {
	byte pins[Movement_PINS_COUNT] = Movement_PINS;
	int i;
	for (i = 0; i < Movement_PINS_COUNT; i++) {
		// Create
		PSensorInfo pinfo = root_new_sensor_info(info, i);
		pinfo->data.bytes[0] = pins[i];
		pinfo->data.bytes[1] = 2; // Will report first value
		pinfo->data.longs[1] = millis();
		pinMode(pins[i], INPUT);
		#ifdef AH_DEBUG
		Serial.print("Movement add: ");
		Serial.println(pins[i]);
		#endif
	}
}

static void movement_on_loop(PSensorInfo pinfo) {
	byte value = digitalRead(pinfo->data.bytes[0]); // Current value
	byte last = pinfo->data.bytes[1];
	unsigned long now = millis();
	unsigned long last_time = pinfo->data.longs[1];
	#ifdef AH_DEBUG
	Serial.print("Movement measure: ");
	Serial.println(value);
	#endif
	if (last == value) {
		// Not changed - skip
		if (value == 1) {
			// Update time when movement detected
			pinfo->data.longs[1] = now;
		}
		return;
	}
	if (last == 1 && now - last_time < Movement_TreshholdMSEC) {
		// Not in movement now, but too early
		return;
	}
	// 1. Changed from 0 to 1 - movement detected
	// 2. Changed from 1 to 0 after treshhold - movement stopped
	pinfo->data.bytes[1] = value;
	pinfo->data.longs[1] = now;
	OutputBuffer output;
	root_new_command(pinfo, &output, CMD_MEASURE);
	root_add_measure(&output, pinfo, 0, value);
	root_send_output(&output);
	// digitalWrite(LED_PIN, value == 1? HIGH: LOW);
}
