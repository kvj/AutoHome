// vim:ft=c
// #define AH_DEBUG
#include "AutoHome.h"

#include "plugin_DHT.cpp"
#include "plugin_Movement.cpp"
#include "plugin_Light.cpp"

#define SENSOR_TYPES 3

SensorTypeInfo sensorTypes[SENSOR_TYPES] = {
	{0, dht_on_create, NULL, NULL, dht_on_measure, NULL},
	{1, movement_on_create, NULL, movement_on_loop, NULL, NULL},
	{2, light_on_create, NULL, light_on_loop, light_on_measure, NULL},
};

PSensorInfo sensors[32];
int sensorsCount = 0;

PSensorInfo root_new_sensor_info(PSensorTypeInfo info, int index) {
	PSensorInfo pinfo = (PSensorInfo)malloc(sizeof(SensorInfo));
	pinfo->type = info->type;
	pinfo->index = index;
	sensors[sensorsCount++] = pinfo;
	#ifdef AH_DEBUG
	Serial.print("root_new_sensor_info: ");
	Serial.println(info->type);
	#endif
	return pinfo;
}

void root_add_measure(POutputBuffer buffer, PSensorInfo pinfo, byte index, byte measure) {
	buffer->buffer[buffer->size++] = pinfo->type;
	buffer->buffer[buffer->size++] = pinfo->index;
	buffer->buffer[buffer->size++] = index;
	buffer->buffer[buffer->size++] = measure;
	#ifdef AH_DEBUG
	Serial.print("Add measure: ");
	Serial.println(measure);
	#endif
}

void byte2chr(byte value, byte *buffer, int index) {
	buffer[index] = (value & 15) + 0x40;
	buffer[index+1] = ((value >> 4) & 15) + 0x40;
}

byte chr2byte(byte *buffer, int index) {
	byte value = (buffer[index] - 0x40) + ((buffer[index+1] - 0x40) << 4);
	return value;
}

bool root_send_output(POutputBuffer buffer) {
	#ifdef AH_DEBUG
	Serial.print("Send output: ");
	Serial.println(buffer->size);
	return true;
	#endif
	byte outbuffer[MAX_OUTPUT*2+3]; // 2 bytes for every byte + size + leading 0
	outbuffer[0] = 0;
	byte2chr(buffer->size, (byte *)&outbuffer, 1);
	int index = 3;
	for (int i = 0; i < buffer->size; i++, index += 2) {
		byte2chr(buffer->buffer[i], (byte *)&outbuffer, index);
	}
	int sent = Serial.write(outbuffer, index);
	Serial.flush();
	return sent == index;
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
	byte buffer[MAX_OUTPUT];
	int size = Serial.read();
	if (size < 0) {
		// Invalid data
		return false;
	}
	if (size>0) {
		// Old style code
		int read = Serial.readBytes((char *)buffer, size);
		if (read != size) {
			// Invalid data
			return false;
		}
	} else {
		// New style
		byte sizeIn[2];
		int read = Serial.readBytes((char *)sizeIn, 2);
		if (read != 2) {
			return false;
		}
		size = chr2byte((byte *)&sizeIn, 0);
		byte bufferIn[2 * MAX_OUTPUT];
		read = Serial.readBytes((char *)&bufferIn, 2*size);
		for (int i = 0; i < size; i++) {
			buffer[i] = chr2byte((byte *)&bufferIn, 2*i);
		}
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
	Serial.println(sensorsCount);
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
		if (NULL != info->on_loop) {
			// OK
			info->on_loop(pinfo);
		}
	}
	root_send_output(&output);
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
