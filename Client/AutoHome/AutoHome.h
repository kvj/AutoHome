#ifndef _AUTOHOME_H_
#define _AUTOHOME_H_

// #define AH_DEBUG

#include <Arduino.h>

#include "AutoHome_config.h"

#define CMD_MEASURE 0
#define MAX_OUTPUT 120

typedef struct {
	int size;
	byte buffer[MAX_OUTPUT];
} OutputBuffer;

typedef OutputBuffer* POutputBuffer;

union Data {
	byte bytes[32];
	int ints[16];
	void *pointers[4];
	unsigned long longs[8];
};

struct Sensor_Info {
	Data data;
	byte type;
	byte index;
};

typedef struct Sensor_Info SensorInfo;
typedef SensorInfo* PSensorInfo;

struct Sensor_Type_Info;
typedef struct Sensor_Type_Info SensorTypeInfo;
typedef SensorTypeInfo* PSensorTypeInfo;

struct Sensor_Type_Info {
	byte type;
	void (* on_create)(PSensorTypeInfo info);
	void (* on_init)(PSensorInfo pinfo);
	void (* on_loop)(PSensorInfo pinfo);
	void (* on_measure)(PSensorInfo pinfo, POutputBuffer buffer);
	void (* on_message)(PSensorInfo pinfo);
};



PSensorInfo root_new_sensor_info(PSensorTypeInfo info, int index);
void root_add_measure(POutputBuffer buffer, PSensorInfo pinfo, byte index, int measure);
void root_new_command(PSensorInfo, POutputBuffer, byte);
bool root_send_output(POutputBuffer);
#endif
