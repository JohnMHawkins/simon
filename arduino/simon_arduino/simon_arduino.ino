//#include <HX711.h>
#include "HX711_ADC.h"
#include <Wire.h>
/*
   This tests i2c communication between the arduino and a r-pi

   Receiving Commands
   it will read an int command from the i2c bus and take action.  The test program only defines one
   cmd, CMD_TEST which will toggle the pin13 LED on the arduino each time it is received

   Sending Data
   It will write a block of data to the i2c bus when the Pi wishes to read it.  The data is an array of bytes that represent
   the normalized weight values of the buttons (normalized = 255 is the max value, 0 = minimum)

   onCommand(cmd) is meant to simulate executing a command.  In this test, it supports one command, toggling the pin13 led

   getWeight(btn) is meant to simulate setting the weight value for a specific button.  In the test, it just sets them to fixed
   values.

   Wiring:
   commectiong an Arduino MEGA2560 to a R-Pi 3 B+:
   arduino pin 20 (SDA) <-> pi 3 (SDA)
   arduino pin 21 (SCL) <-> pi 5 (SCL)
   arduino GND <-> pi GND

*/

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
#define CMD_ZERO   2


//center button

byte buttonPin = 53;
bool CenterPressed = false;

void setupCenterButton()
{
  pinMode(buttonPin, INPUT);
}

void updateCenterButton()
{
  CenterPressed = digitalRead(buttonPin);

}

// new button

const byte dataPins[4][4] = {{4, 5, 6, 7}, {14, 22, 24, 25}, {A0, A1, A2, A3}, {A4, A9, A10, A11}} ; //{{4, 5, 6, 7}, {14, 27, 28, 29}, {A0, A1, A2, A3}, {A8, A9, A10, A11}};
const byte clockPins[4][4] = {{3, 3, 3, 3}, {2, 2, 2, 2}, {3, 3, 3, 3}, {2, 2, 2, 2}};
//HX711 loadCell[4][4];




#define mWeight 600





//long maxWeight[4] = {mWeight, mWeight, mWeight, mWeight};

HX711_ADC lc_0_0(dataPins[0][0], clockPins[0][0]);
HX711_ADC lc_0_1(dataPins[0][1], clockPins[0][1]);
//HX711_ADC lc_0_2(dataPins[0][2], clockPins[0][2]);
//HX711_ADC lc_0_3(dataPins[0][3], clockPins[0][3]);

HX711_ADC lc_1_0(dataPins[1][0], clockPins[1][0]);
HX711_ADC lc_1_1(dataPins[1][1], clockPins[1][1]);
//HX711_ADC lc_1_2(dataPins[1][2], clockPins[1][2]);
//HX711_ADC lc_1_3(dataPins[1][3], clockPins[1][3]);

HX711_ADC lc_2_0(dataPins[2][0], clockPins[2][0]);
HX711_ADC lc_2_1(dataPins[2][1], clockPins[2][1]);
//HX711_ADC lc_2_2(dataPins[2][2], clockPins[2][2]);
//HX711_ADC lc_2_3(dataPins[2][3], clockPins[2][3]);

HX711_ADC lc_3_0(dataPins[3][0], clockPins[3][0]);
HX711_ADC lc_3_1(dataPins[3][1], clockPins[3][1]);
//HX711_ADC lc_3_2(dataPins[3][2], clockPins[3][2]);
//HX711_ADC lc_3_3(dataPins[3][3], clockPins[3][3]);


void setupCells() {

  lc_0_0.begin();
  lc_0_1.begin();
//  lc_0_2.begin();
//  lc_0_3.begin();

  lc_1_0.begin();
  lc_1_1.begin();
//  lc_1_2.begin();
//  lc_1_3.begin();

  lc_2_0.begin();
  lc_2_1.begin();
//  lc_2_2.begin();
//  lc_2_3.begin();

  lc_3_0.begin();
  lc_3_1.begin();
//  lc_3_2.begin();
//  lc_3_3.begin();

  lc_0_0.start(2000);
  lc_0_1.start(2000);
//  lc_0_2.start(2000);
//  lc_0_3.start(2000);

  lc_1_0.start(2000);
  lc_1_1.start(2000);
//  lc_1_2.start(2000);
//  lc_1_3.start(2000);

  lc_2_0.start(2000);
  lc_2_1.start(2000);
//  lc_2_2.start(2000);
//  lc_2_3.start(2000);

  lc_3_0.start(2000);
  lc_3_1.start(2000);
//  lc_3_2.start(2000);
//  lc_3_3.start(2000);

  //  lc_0_0.startMultiple(2000);
  //  lc_0_1.startMultiple(2000);
  //  lc_0_2.startMultiple(2000);
  //  lc_0_3.startMultiple(2000);
  //
  //  lc_1_0.startMultiple(2000);
  //  lc_1_1.startMultiple(2000);
  //  lc_1_2.startMultiple(2000);
  //  lc_1_3.startMultiple(2000);
  //
  //  lc_2_0.startMultiple(2000);
  //  lc_2_1.startMultiple(2000);
  //  lc_2_2.startMultiple(2000);
  //  lc_2_3.startMultiple(2000);
  //
  //  lc_3_0.startMultiple(2000);
  //  lc_3_1.startMultiple(2000);
  //  lc_3_2.startMultiple(2000);
  //  lc_3_3.startMultiple(2000);

  const int cf = 696;

  lc_0_0.setCalFactor(cf);
  lc_0_1.setCalFactor(cf);
//  lc_0_2.setCalFactor(cf);
//  lc_0_3.setCalFactor(cf);

  lc_1_0.setCalFactor(cf);
  lc_1_1.setCalFactor(cf);
//  lc_1_2.setCalFactor(cf);
//  lc_1_3.setCalFactor(cf);

  lc_2_0.setCalFactor(cf);
  lc_2_1.setCalFactor(cf);
//  lc_2_2.setCalFactor(cf);
//  lc_2_3.setCalFactor(cf);

  lc_3_0.setCalFactor(cf);
  lc_3_1.setCalFactor(cf);
//  lc_3_2.setCalFactor(cf);
//  lc_3_3.setCalFactor(cf);


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

  lc_0_0.tare();
  lc_0_1.tare();
//  lc_0_2.tare();
//  lc_0_3.tare();

  lc_1_0.tare();
  lc_1_1.tare();
//  lc_1_2.tare();
//  lc_1_3.tare();

  lc_2_0.tare();
  lc_2_1.tare();
//  lc_2_2.tare();
//  lc_2_3.tare();

  lc_3_0.tare();
  lc_3_1.tare();
//  lc_3_2.tare();
//  lc_3_3.tare();



  //  lc_0_0.tareNoDelay();
  //  lc_0_1.tareNoDelay();
  //  lc_0_2.tareNoDelay();
  //  lc_0_3.tareNoDelay();
  //
  //  lc_1_0.tareNoDelay();
  //  lc_1_1.tareNoDelay();
  //  lc_1_2.tareNoDelay();
  //  lc_1_3.tareNoDelay();
  //
  //  lc_2_0.tareNoDelay();
  //  lc_2_1.tareNoDelay();
  //  lc_2_2.tareNoDelay();
  //  lc_2_3.tareNoDelay();
  //
  //  lc_3_0.tareNoDelay();
  //  lc_3_1.tareNoDelay();
  //  lc_3_2.tareNoDelay();
  //  lc_3_3.tareNoDelay();


  /*
    for (int i = 0; i < 4; i++)
    {
    for (int j = 0; j < 4; j++)
    {
      loadCell[i][j].tare();
    }
    //Serial.println(maxWeight[i]);
    maxWeight[i] = mWeight;
    }
  */

}

byte getCellWeight(int Button)
{

  float weight = 0;

  switch (Button) {
    case BTN_RED:
      weight = weight + lc_0_0.update();
      weight = weight + lc_0_1.update();
//      weight = weight + lc_0_2.update();
//      weight = weight + lc_0_3.update();
      //
      weight = weight + lc_0_0.getData();
      weight = weight + lc_0_1.getData();
//      weight = weight + lc_0_2.getData();
//      weight = weight + lc_0_3.getData();
      break;

    case BTN_GREEN:
      weight = weight + lc_1_0.update();
      weight = weight + lc_1_1.update();
//      weight = weight + lc_1_2.update();
//      weight = weight + lc_1_3.update();
      //
      weight = weight + lc_1_0.getData();
      weight = weight + lc_1_1.getData();
//      weight = weight + lc_1_2.getData();
//      weight = weight + lc_1_3.getData();
      break;

    case BTN_BLUE:
      weight = weight + lc_2_0.update();
      weight = weight + lc_2_1.update();
//      weight = weight + lc_2_2.update();
//      weight = weight + lc_2_3.update();
      //
      weight = weight + lc_2_0.getData();
      weight = weight + lc_2_1.getData();
//      weight = weight + lc_2_2.getData();
//      weight = weight + lc_2_3.getData();
      break;

    case BTN_YELLOW:
      weight = weight + lc_3_0.update();
      weight = weight + lc_3_1.update();
//      weight = weight + lc_3_2.update();
//      weight = weight + lc_3_3.update();
      //
      weight = weight + lc_3_0.getData();
      weight = weight + lc_3_1.getData();
//      weight = weight + lc_3_2.getData();
//      weight = weight + lc_3_3.getData();
      break;

    default:
      weight = -1;
      break;

  }

//  weight = weight / 4.0;
  if(weight > mWeight) weight = mWeight;
  byte normalWeight = (byte)(255 * weight / (4.0 * mWeight));

  Serial.println("Button: [" + (String)Button + "]: " + (String)weight);

  return normalWeight;

  /*

    int button = Button - 1;

    long weight = 0;


    for (int i = 0; i < 4; i++)
    {
    //Serial.println("getting load cell data cell: " + (String)i);
    long w = loadCell[button][i].get_value(1);
    if(button == BTN_YELLOW )
    {
    //Serial.println("Cell [" + (String)button + "][" + (String)i + "]: " + (String)w);
    }
    if (w < 100) w = 0;
    weight += w;
    //if(button == BTN_YELLOW )
    //{
    //Serial.println("Cell [" + (String)button + "][" + (String)i + "]: " + (String)w);
    //}
    }



    weight /= 4;

    if (weight > mWeight) weight = mWeight;

    byte normalWeight = (byte)((double)weight / (mWeight) * 255);

    //if(button == 2)
    //Serial.println(normalWeight);

    return normalWeight;
  */
}

// weights are byte values 0-255
byte weights[] = {0, 0, 0, 0, 0};
int state = 0;

void setup() {
  // put your setup code here, to run once:
  //pinMode(13, OUTPUT);


  // initialize i2c slave
  Wire.begin(SLAVE_ADDRESS);

  // define i2c callbacks
  Wire.onReceive(receiveData);
  Wire.onRequest(sendData);


  //init buttons

  Serial.begin(9600);
  //Serial.println("Initing...");

  setupCells();
  setupCenterButton();

  //Serial.println("Buttons Inited!");

  zeroCells();

  //Serial.println("Finished!");



}

void loop() {
  Wire.begin(SLAVE_ADDRESS);
  // put your main code here, to run repeatedly:
  updateCenterButton();


  weights[0] = getWeight(0);
  weights[1] = getWeight(1);
  weights[2] = getWeight(2);
  weights[3] = getWeight(3);
  weights[4] = getWeight(4);


  //Serial.println("\n\nGetting buttons:");
  for (int i = 0; i < 5; i++)
  {
    //Serial.println(weights[i]);
  }

  yield();
}

// received a command from the Pi over the i2c bus
void receiveData(int byteCount) {
  int cmd;
  while (Wire.available()) {
    cmd = Wire.read();
    onCommand(cmd);
  }
}

// send data to the Pi over the i2c bus
void sendData() {
  Wire.write(weights, sizeof(weights));

}

void onCommand(int cmd) {

  switch (cmd) {
    case CMD_TEST:
      if (state == 0) {
        digitalWrite(13, HIGH); // set the LED on
        state = 1;
      } else {
        digitalWrite(13, LOW);  // set the LED off
        state = 0;
      }
      break;
    case CMD_ZERO:
      zeroCells();
      break;
    default:
      break;
  }
}

// This method returns the weight for the specified button as a byte (0-255)
byte getWeight(int btn) {
  byte d;
  switch (btn) {
    case BTN_CENTER:
      if (CenterPressed)
      {
        return 100;
      }
      else
      {
        return 0;
      }
      break;
    case BTN_RED:
      //return 255;
      d = getCellWeight(BTN_RED);
      if (d > 0)
      {
        return d;
      }
      else
      {
        return 0;
      }
      break;
    case BTN_GREEN:
      //return 0;
      d = getCellWeight(BTN_GREEN);
      if (d > 0)
      {
        return d;
      }
      else
      {
        return 0;
      }
      break;
    case BTN_BLUE:
      //return 0;
      d = getCellWeight(BTN_BLUE);
      if (d > 0)
      {
        return d;
      }
      else
      {
        return 0;
      }
      break;
    case BTN_YELLOW:
      // return 0;
      d = getCellWeight(BTN_YELLOW);
      if (d > 0)
      {
        return d;
      }
      else
      {
        return 0;
      }
      break;

  }

  return 0;
}
