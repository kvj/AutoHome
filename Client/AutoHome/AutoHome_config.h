#ifndef _AUTOHOME_CONFIG_H_
#define _AUTOHOME_CONFIG_H_

#define Movement_TreshholdMSEC 60 * 1000
#define Light_Treshhold 10

#define LED_PIN 13

#define DEVICE_CONFIG_1

#ifdef DEVICE_CONFIG_0
	#define DHT_PINS_COUNT 1
	#define DHT_PINS {6}
	#define Movement_PINS_COUNT 1
	#define Movement_PINS {4}
	#define Light_PINS_COUNT 1
	#define Light_PINS {A0}
#endif
#ifdef DEVICE_CONFIG_1
	#define DHT_PINS_COUNT 1
	#define DHT_PINS {6}
	#define Movement_PINS_COUNT 1
	#define Movement_PINS {4}
	#define Light_PINS_COUNT 1
	#define Light_PINS {A0}
#endif

#endif
