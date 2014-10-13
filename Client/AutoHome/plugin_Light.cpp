#include "AutoHome.h"
#include <math.h>

static void light_on_create(PSensorTypeInfo info) {
	byte pins[Light_PINS_COUNT] = Light_PINS;
	int i;
	for (i = 0; i < Light_PINS_COUNT; i++) {
		// Create
		PSensorInfo pinfo = root_new_sensor_info(info, i);
		pinfo->data.bytes[0] = pins[i];
		pinfo->data.bytes[1] = 0; // Initial light
		pinfo->data.longs[1] = millis();
	}
}

static void light_on_loop(PSensorInfo pinfo) {
	unsigned long now = millis();
	unsigned long last_time = pinfo->data.longs[1];
	byte current = (analogRead(pinfo->data.bytes[0]) / 4);
	if (abs(pinfo->data.bytes[1] - current) <= Light_Treshhold) {
		return; // Nothing to send yet
	}
	pinfo->data.bytes[1] = current;
	pinfo->data.longs[1] = now;

	OutputBuffer output;
	root_new_command(pinfo, &output, CMD_MEASURE);
	root_add_measure(&output, pinfo, 0, current);
	root_send_output(&output);
}
