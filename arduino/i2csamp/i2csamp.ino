#include <Wire.h>
#include "HX711.h"
#include "HX711_ADC.h"


/*
 * This tests i2c communication between the arduino and a r-pi
 * 
 * Receiving Commands
 * it will read an int command from the i2c bus and take action.  The test program only defines one
 * cmd, CMD_TEST which will toggle the pin13 LED on the arduino each time it is received
 * 
 * Sending Data
 * It will write a block of data to the i2c bus when the Pi wishes to read it.  The data is an array of bytes that represent 
 * the normalized weight values of the buttons (normalized = 255 is the max value, 0 = minimum)
 * 
 * onCommand(cmd) is meant to simulate executing a command.  In this test, it supports one command, toggling the pin13 led
 * 
 * getWeight(btn) is meant to simulate setting the weight value for a specific button.  In the test, it just sets them to fixed
 * values.
 * 
 * Wiring:
 * commectiong an Arduino MEGA2560 to a R-Pi 3 B+:
 * arduino pin 20 (SDA) <-> pi 3 (SDA)
 * arduino pin 21 (SCL) <-> pi 5 (SCL)
 * arduino GND <-> pi GND
 * 
 */



// HX711 circuit wiring
const int LOADCELL_DOUT_PIN = 2;
const int LOADCELL_SCK_PIN = 3;

//HX711 scale;
HX711_ADC LoadCell_1(LOADCELL_DOUT_PIN, LOADCELL_SCK_PIN); //HX711 1

/*
const byte dataPins[4][4] = {{4, 5, 6, 6}, {14, 22, 24, 25}, {A0, A1, A2, A3}, {A4, A9, A10, A11}} ; //{{4, 5, 6, 7}, {14, 27, 28, 29}, {A0, A1, A2, A3}, {A8, A9, A10, A11}};
const byte clockPins[4][4] = {{3, 3, 3, 3}, {2, 2, 2, 2}, {3, 3, 3, 3}, {2, 2, 2, 2}};
HX711 loadCell[4][4];
#define mWeight 160000
*/

void setupCells() {

  //scale.begin(LOADCELL_DOUT_PIN, LOADCELL_SCK_PIN);
  LoadCell_1.start(200);
  
  /*
  for (int i = 0; i < 4; i++)
  {
    for (int j = 0; j < 4; j++)
    {
      //Serial.println("test:");
      loadCell[i][j].begin(dataPins[i][j], clockPins[i][j]);
      //loadCell[i][j].tare();
    }
  }
  */
}

void zeroCells() {

  LoadCell_1.tareNoDelay();
  //if (scale.wait_ready_timeout(1000)) {
    //scale.tare();
  //}
  /*
  for (int i = 0; i < 4; i++)
  {
    for (int j = 0; j < 4; j++)
    {
      if (loadCell[i][j].wait_ready_timeout(1000)) {
        //Serial.print("Tare the cells");
        //loadCell[i][j].tare();
      }
    }
  }
  */
}



byte xgetCellWeight(int Button)
{
  LoadCell_1.getData();
  //scale.read();
  return 0;
  /*
  int button = Button - 1;

  long weight = 0;

  //Serial.println("getting load cell data cell: " + (String)button);

  for (int i = 0; i < 4; i++)
  {
    if (loadCell[button][i].wait_ready_timeout(1000)) {
      long w = loadCell[button][i].read();
      //long w = loadCell[button][i].get_value(1);
      //long w = 1;
      weight = weight + w;
    }
  }
 
  weight /= 4;

  if (weight > mWeight) weight = mWeight;

  byte normalWeight = (byte)((double)weight / (mWeight) * 255);
  return normalWeight;
  */
}


// this has to match the ardAdr param in the py program
#define SLAVE_ADDRESS 0x03

// buttons
#define BTN_CENTER  0
#define BTN_RED     1
#define BTN_GREEN   2
#define BTN_BLUE    3
#define BTN_YELLOW  4

// commands
#define CMD_TEST   1

// weights are byte values 0-255
byte weights[] = {0, 0, 0, 0, 0};
int state = 0;

char rxChar = 0;

int debugcounter = 0;

byte getCellWeight(int Button) {
  return xgetCellWeight(Button);
  //return weights[Button];
}


void setup() {
  // put your setup code here, to run once:
  pinMode(13, OUTPUT);

  // define serial to computer
  Serial.begin(9600);
  Serial.flush();

  // initialize i2c slave
  Wire.begin(SLAVE_ADDRESS);

  // define i2c callbacks
  //Wire.onReceive(receiveData);
  Wire.onRequest(sendData);

  //scale.begin(LOADCELL_DOUT_PIN, LOADCELL_SCK_PIN);
  Serial.print("called begin on HX711");
  setupCells();

}

void simulateButtons() {
  for (int k =0; k < 1; k++ ) {
    int j = random(10000, 99999);
    int idx = 0;
    for (int i = 0; i < j; i++ ) {
      if (weights[idx] >= 255) {
        weights[idx] = 0;
        idx = idx + 1;
        if (idx > 4 ){
          idx = 0;
        }
      } else {
        weights[idx] = weights[idx] + 1;
      }
    }
    
  }
}



void loop() {
  // put your main code here, to run repeatedly:
  delay(100);
  //Serial.println("Hello");
  simulateButtons();
  
  if (Serial.available() > 0 ){
    rxChar = Serial.read();
    //Serial.flush();
    switch (rxChar) {
      case 'c':
        for (int i = 0; i < 5; i++ ) {
            weights[i] = 0;
        }
        weights[0] = 100;
        Wire.begin(SLAVE_ADDRESS);
        Wire.onRequest(sendData);
        break;

      case '1':
        for (int i = 0; i < 5; i++ ) {
            weights[i] = 0;
        }
        weights[1] = 100;
        break;

      case '2':
        for (int i = 0; i < 5; i++ ) {
            weights[i] = 0;
        }
        weights[2] = 100;
        break;

      case '3':
        for (int i = 0; i < 5; i++ ) {
          weights[i] = 0;
        }
        weights[3] = 100;
        break;

      case '4':
        for (int i = 0; i < 5; i++ ) {
          weights[i] = 0;
        }
        weights[4] = 100;
        break;

    }
    

  
  }
  

  zeroCells();
  for ( int k = 1; k < 5; k++) {
    getCellWeight(k);
  }
  /*
  if (scale.is_ready()) {
    long reading = scale.read();
    Serial.print("HX711 reading: ");
    Serial.println(reading);
  } else {
    Serial.println("HX711 not found.");
  }
  */

  /*
  weights[0] = getWeight(0);
  weights[1] = getWeight(1);
  weights[2] = getWeight(2);
  weights[3] = getWeight(3);
  weights[4] = getWeight(4);
  */
}

// received a command from the Pi over the i2c bus
void receiveData(int byteCount) {
int cmd;
  while(Wire.available()) {
    cmd = Wire.read();
    onCommand(cmd);
  }
}

byte noweights[] = {0, 0, 0, 0, 0};

// send data to the Pi over the i2c bus
void sendData() {
  //Serial.println(weights[0] );
  Wire.write(noweights, sizeof(noweights));
  
}

void onCommand(int cmd) {

  
  switch(cmd){
    case CMD_TEST:
      if (state == 0) {
        digitalWrite(13, HIGH); // set the LED on
        state = 1;        
      } else {
        digitalWrite(13, LOW);  // set the LED off
        state = 0;
      }
      break;
    default:
      break;    
  }
  
}

// This method returns the weight for the specified button as a byte (0-255)
byte getWeight(int btn) {
  switch(btn) {
    case BTN_CENTER:
      return 0;
      break;
    case BTN_RED:
      return 120;
      break;
    case BTN_GREEN:
      return 130;
      break;
    case BTN_BLUE:
      return 140;
      break;
    case BTN_YELLOW:
      return 150;
      break;
    
  }

  return 0;
}
