// vim:ft=c
#include "AutoHome.h"

#include "plugin_DHT.cpp"
#include "plugin_Movement.cpp"
#include "plugin_Light.cpp"

#define SENSOR_TYPES 3

SensorTypeInfo sensorTypes[SENSOR_TYPES] = {
	{0, dht_on_create, NULL, NULL, dht_on_measure, NULL},
	{1, movement_on_create, NULL, movement_on_loop, NULL, NULL},
	{2, light_on_create, NULL, light_on_loop, NULL, NULL},
};

PSensorInfo sensors[32];
int sensorsCount = 0;

PSensorInfo root_new_sensor_info(PSensorTypeInfo info, int index) {
	PSensorInfo pinfo = (PSensorInfo)malloc(sizeof(SensorInfo));
	pinfo->type = info->type;
	pinfo->index = index;
	sensors[sensorsCount++] = pinfo;
	return pinfo;
}

void root_add_measure(POutputBuffer buffer, PSensorInfo pinfo, byte index, byte measure) {
	buffer->buffer[buffer->size++] = pinfo->type;
	buffer->buffer[buffer->size++] = pinfo->index;
	buffer->buffer[buffer->size++] = index;
	buffer->buffer[buffer->size++] = measure;
}

bool root_send_output(POutputBuffer buffer) {
	Serial.write(buffer->size);
	int sent = Serial.write(buffer->buffer, buffer->size);
	Serial.flush();
	return sent == buffer->size;
}

void root_new_command(PSensorInfo pinfo, POutputBuffer output, byte command) {
	output->buffer[0] = command;
	output->size = 1;
}

bool root_process_input() {
	if (Serial.available() <= 0) {
		// No data
		return false;
	}
	byte buffer[256];
	int size = Serial.read();
	if (size <= 0) {
		// Invalid data
		return false;
	}
	int read = Serial.readBytes((char *)buffer, size);
	if (read != size) {
		// Invalid data
		return false;
	}
	// Data OK
	if (buffer[0] == CMD_MEASURE) {
		// Make measurements
		OutputBuffer output;
		output.buffer[0] = CMD_MEASURE;
		output.size = 1;
		int i;
		for (i = 0; i < sensorsCount; i++) {
			PSensorInfo pinfo = sensors[i];
			PSensorTypeInfo info = &sensorTypes[pinfo->type];
			if (NULL != info->on_measure) {
				// OK
				info->on_measure(pinfo, &output);
			}
		}
		return root_send_output(&output);
	}
	return true;
}

void setup() {
	Serial.begin(9600);
	pinMode(LED_PIN, OUTPUT);
	digitalWrite(LED_PIN, LOW);
	int i;
	for (i = 0; i < SENSOR_TYPES; i++) {
		PSensorTypeInfo info = &sensorTypes[i];
		if (info->on_create != NULL) {
			info->on_create(info);
		}
	}
	#ifdef AH_DEBUG
	Serial.print("Debug mode: ");
	Serial.println(sensorsCount);
	#endif
}

void loop() {
	#ifdef AH_DEBUG
	delay(3000);
	Serial.println("One cycle:");
	OutputBuffer output;
	output.buffer[0] = CMD_MEASURE;
	output.size = 1;
	int i;
	for (i = 0; i < sensorsCount; i++) {
		PSensorInfo pinfo = sensors[i];
		PSensorTypeInfo info = &sensorTypes[pinfo->type];
		if (NULL != info->on_measure) {
			// OK
			info->on_measure(pinfo, &output);
		}
	}
	return;

	#endif
	bool read_result = root_process_input();
	if (!read_result) {
		// Do loop
		int i;
		for (i = 0; i < sensorsCount; i++) {
			PSensorInfo pinfo = sensors[i];
			PSensorTypeInfo info = &sensorTypes[pinfo->type];
			if (NULL != info->on_loop) {
				// OK
				info->on_loop(pinfo);
			}
		}
		delay(100);
	}
}
