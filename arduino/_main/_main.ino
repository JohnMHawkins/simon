#include <HX711.h>
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

const byte dataPins[4][4] = {{4, 5, 6, 6}, {14, 22, 24, 25}, {A0, A1, A2, A3}, {A4, A9, A10, A11}} ; //{{4, 5, 6, 7}, {14, 27, 28, 29}, {A0, A1, A2, A3}, {A8, A9, A10, A11}};
const byte clockPins[4][4] = {{3, 3, 3, 3}, {3, 3, 3, 3}, {3, 3, 3, 3}, {3, 3, 3, 3}};
HX711 loadCell[4][4];
#define mWeight 160000
long maxWeight[4] = {mWeight, mWeight, mWeight, mWeight};

void setupCells() {
  for (int i = 0; i < 4; i++)
  {
    for (int j = 0; j < 4; j++)
    {
      loadCell[i][j].begin(dataPins[i][j], clockPins[i][j]);
      //loadCell[i][j].tare();
    }
  }
}

void zeroCells() {
  for (int i = 0; i < 4; i++)
  {
    for (int j = 0; j < 4; j++)
    {
      loadCell[i][j].tare();
    }
    //Serial.println(maxWeight[i]);
    maxWeight[i] = mWeight;
  }

}

byte getCellWeight(int Button)
{
  int button = Button - 1;

  long weight = 0;


  for (int i = 0; i < 4; i++)
  {
    //Serial.println("getting load cell data cell: " + (String)i);
    long w = loadCell[button][i].get_value(1);
    if (w < 100) w = 0;
    weight += w;
    //if(button == 1 )
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

  //Serial.begin(9600);
  //Serial.println("Initing...");

  setupCells();

  setupCenterButton();
  zeroCells();

  //Serial.println("Buttins Inited!");

  //Serial.println("Finished!");



}

void loop() {
  // put your main code here, to run repeatedly:
  updateCenterButton();


  weights[0] = getWeight(0);
  weights[1] = getWeight(1);
  weights[2] = getWeight(2);
  weights[3] = getWeight(3);
  weights[4] = getWeight(4);

  //sendData();
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
      if(d > 0)
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
      if(d > 0)
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
      if(d > 0)
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
      if(d > 0)
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
