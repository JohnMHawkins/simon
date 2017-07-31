#include <bitswap.h>
#include <chipsets.h>
#include <color.h>
#include <colorpalettes.h>
#include <colorutils.h>
#include <controller.h>
#include <cpp_compat.h>
#include <dmx.h>
#include <FastLED.h>
#include <fastled_config.h>
#include <fastled_delay.h>
#include <fastled_progmem.h>
#include <fastpin.h>
#include <fastspi.h>
#include <fastspi_bitbang.h>
#include <fastspi_dma.h>
#include <fastspi_nop.h>
#include <fastspi_ref.h>
#include <fastspi_types.h>
#include <hsv2rgb.h>
#include <led_sysdefs.h>
#include <lib8tion.h>
#include <noise.h>
#include <pixelset.h>
#include <pixeltypes.h>
#include <platforms.h>
#include <power_mgt.h>

#include <FastLED.h>

// vars for serial communication
String inputString = "";
String simonCommand = "";
boolean stringComplete = false; 
String deviceName = "unknown";
String deviceInput = "unknown";

//Arduino Identification Pins (set to ground)
int PIN_CENTER = 51;

int PIN_BLUEGREEN = 52;
int PIN_REDYELLOW = 53;

int PIN_BLUE = 50;
int PIN_RED = 49;
int PIN_GREEN = 48;
int PIN_YELLOW = 47;

// global vars for simon
String simonMode = "NONE";

void setup() {
  // put your setup code here, to run once:
  Serial.begin (115200);
  Serial.setTimeout (50);

  inputString.reserve(200);

  //setup arduino identification pins
  setupIDPin();
}

void setupIDPin() {
  pinMode(PIN_CENTER,INPUT);
  pinMode(PIN_BLUEGREEN,INPUT);
  pinMode(PIN_REDYELLOW,INPUT);  
  pinMode(PIN_BLUE,INPUT);
  pinMode(PIN_GREEN,INPUT);
  pinMode(PIN_RED,INPUT);  
  pinMode(PIN_YELLOW,INPUT);  

  digitalWrite(PIN_CENTER, HIGH);
  digitalWrite(PIN_BLUEGREEN, HIGH);
  digitalWrite(PIN_REDYELLOW, HIGH);
  digitalWrite(PIN_BLUE, HIGH);
  digitalWrite(PIN_GREEN, HIGH);
  digitalWrite(PIN_RED, HIGH);
  digitalWrite(PIN_YELLOW, HIGH);    
}


void loop() {
  // put your main code here, to run repeatedly:

  
  // parse the string when a newline arrives:
  if (stringComplete) {

    parseCommand();

  }
  
}


// 
// helper function to output to serial port
//
void simonOutput (String outputString) {
    Serial.print(deviceName);
    Serial.print("["+deviceInput+"]");
    Serial.print(":");
    Serial.println(outputString);         
}

//
// function to compare our inputString to possible commands
// add more commands and calls to funcitons here
//
void parseCommand () {

  //our giant if statement to compare all commands

  //assign device nmae
  if (inputString.substring(0,4) == "NAME" ) {
    //assign device nmae
    simonAssignDeviceName();
 
  } else {
    //get deviceInputName and process command
    deviceInput = getValue(inputString, ':', 0);    
    simonCommand = getValue(inputString, ':', 1);
    //Serial.println(deviceInput + "  " + simonCommand);

    processSimonCommand();    
  }
 
  // clear inputString, deviceInput & commandString
  inputString = "";
  deviceInput = "";
  simonCommand = "";
  stringComplete = false;

}


void processSimonCommand() {
  
//  if (inputString.substring(0,10) == "FLASHCOLOR" ) {
  if (simonCommand == "FLASHCOLOR") {
    //simonOutput(inputString);
    String colorRGB = getValue(inputString, ':', 2);
    
    String colorRed = getValue(colorRGB, '-', 0);
    String colorGreen = getValue(colorRGB, '-', 1);
    String colorBlue = getValue(colorRGB, '-', 2);
    
    String colorTime = getValue(inputString, ':', 3);
    simonOutput(deviceInput + ":" + colorRed + ":"+ colorGreen + ":" + colorBlue + " TIME:" + colorTime);
  
  } else if (simonCommand == "GS_ATTRACT") {
    simonOutput (simonCommand);
    //Call your attract function here
         
  } else if (simonCommand == "GS_COMPUTER") {
    simonOutput (simonCommand);
    //Call compiter function here

  } else if (simonCommand == "GS_TIMER") {
    simonOutput (simonCommand);
    String colorRGB = getValue(inputString, ':', 2);
    
    String colorRed = getValue(colorRGB, '-', 0);
    String colorGreen = getValue(colorRGB, '-', 1);
    String colorBlue = getValue(colorRGB, '-', 2);
    
    String colorTime = getValue(inputString, ':', 3);
    simonOutput(deviceInput + ":" + colorRed + ":"+ colorGreen + ":" + colorBlue + " TIME:" + colorTime);


  } else if (simonCommand == "GS_END_GAME") {
    simonOutput (simonCommand);
    //Call your timer function here
  
  } else if (simonCommand == "READ_BUTTONS") {
    //create test button data, replace with real button function
    simonOutput ("BTTN:1000");     
  
  } else {
    simonOutput ("NOCMD:" + simonCommand);
  }
 
}

//
// serialEvent grabs all input from serial port and creates an inputString.
//
void serialEvent () {
  // called when serial port has data
  while (Serial.available() && !stringComplete) {
    char inChar = (char)Serial.read();

    // if newline, setflag so main loop can do something about it
    if (inChar == '\n') {
      stringComplete = true;
    } else {
      //add it to inputString
      inputString += inChar;
    }
  }

}


/*
 * takes a string and separates it based on a given character and returns The item between the separating character 
 */
String getValue(String data, char separator, int index)
{
    int found = 0;
    int strIndex[] = { 0, -1 };
    int maxIndex = data.length() - 1;

    for (int i = 0; i <= maxIndex && found <= index; i++) {
        if (data.charAt(i) == separator || i == maxIndex) {
            found++;
            strIndex[0] = strIndex[1] + 1;
            strIndex[1] = (i == maxIndex) ? i+1 : i;
        }
    }
    return found > index ? data.substring(strIndex[0], strIndex[1]) : "";
}


/*
 * all functions related to parse commands
 * 
 */



void simonAssignDeviceName() {
      
    //deviceName = getValue(inputString, ':', 1);

    int valCENTER = digitalRead(PIN_CENTER);
    int valBLUEGREEN = digitalRead(PIN_BLUEGREEN);
    int valREDYELLOW = digitalRead(PIN_REDYELLOW);
    int valBLUE = digitalRead(PIN_BLUE);
    int valGREEN = digitalRead(PIN_GREEN);
    int valRED = digitalRead(PIN_RED);
    int valYELLOW = digitalRead(PIN_YELLOW);

    if (valCENTER == LOW) {
      deviceName = "CENTER";
    } else if (valBLUEGREEN == LOW) {
      deviceName = "BLUEGREEN";      
    } else if (valREDYELLOW == LOW) {
      deviceName = "REDYELLOW";      
    } else if (valBLUE == LOW) {
      deviceName = "BLUE";
    } else if (valGREEN == LOW) {
      deviceName = "GREEN";      
    } else if (valRED == LOW) {
      deviceName = "RED";      
    } else if (valYELLOW == LOW) {
      deviceName = "YELLOW";            
    } else {
      deviceName = "ERROR:UNKNOWN";
    }

    Serial.print("NAME:");
    Serial.println(deviceName);      
 
}

