#include <FastLED.h>

// vars for serial communication
String inputString = "";
boolean stringComplete = false; 
String deviceName = "unknown";

// global vars for simon
String simonMode = "NONE";

void setup() {
  // put your setup code here, to run once:
  Serial.begin (115200);
  Serial.setTimeout (50);

  inputString.reserve(200);
  
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
    Serial.print(":");
    Serial.println(outputString);         
}

//
// function to compare our inputString to possible commands
// add more commands and calls to funcitons here
//
void parseCommand () {

  //our giant if statement to compare all commands

  if (inputString.substring(0,4) == "NAME" ) {
    simonAssignDeviceName();
 
  } else if (inputString.substring(0,10) == "FLASHCOLOR" ) {
    //simonOutput(inputString);
    String colorRGB = getValue(inputString, ':', 1);
    
    String colorRed = getValue(colorRGB, '-', 0);
    String colorGreen = getValue(colorRGB, '-', 1);
    String colorBlue = getValue(colorRGB, '-', 2);
    
    String colorTime = getValue(inputString, ':', 2);
    simonOutput("COLOR:" + colorRed + "::"+ colorGreen + "::" + colorBlue + " TIME:" + colorTime);
  
  } else if (inputString == "GS_ATTRACT") {
    simonOutput (inputString);
    //Call your attract function here
         
  } else if (inputString == "GS_COMPUTER") {
    simonOutput (inputString);
    //Call compiter function here

  } else if (inputString == "GS_TIMER") {
    simonOutput (inputString);
    //Call your timer function here
  
  } else if (inputString == "READ_BUTTONS") {
    //create test button data, replace with real button function
    simonOutput ("BTTN:1000");     
  
  } else {
    simonOutput ("NOCMD:" + inputString);
  }

  // clear inputString 
  inputString = "";
  stringComplete = false;

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
      
    deviceName = getValue(inputString, ':', 1);

    Serial.print("DEVICE ASSIGNED:");
    Serial.println(deviceName);

}

