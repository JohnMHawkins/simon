#include <Arduino.h>

#include <TimerThree.h>
#include <FastLED.h>
#include "HX711.h"

//enums
enum CallBack {FLASHCOLOR, COUNTDOWN, UNDEF};

enum State {RAINBOW, SOLID, COUNTING};

//init leds
//light settings 1
#define DATA_PIN1    2
#define CLK_PIN1     3
#define LED_TYPE1    APA102
#define COLOR_ORDER1 BGR
#define NUM_LEDS1    144

#define BRIGHTNESS 255

//light settings 2
#define DATA_PIN2    4
#define CLK_PIN2     5
#define LED_TYPE2    APA102
#define COLOR_ORDER2 BGR
#define NUM_LEDS2    144

CRGB leds1[NUM_LEDS1];
CRGB leds2[NUM_LEDS2];

State ledState[] = {SOLID, SOLID};

//event manager variables
int timer[] = { -1, -1};
CallBack timerCallback[] = {UNDEF, UNDEF};

//Color
CRGB ledColor[] = {CRGB::Black, CRGB::Black};

//FLASHCOLOR variables
int flashAmmount[] = {0, 0};
int timePerFlash[] = {0, 0};
int timeBetweenFlash[] = {0, 0};
bool isFlashLit[] = {false, false};

//RAINBOW variables
uint8_t hue[] = {0, 0};

//Countdown Variables
float countdownAmmount[] = {0, 0};
float PercentagePerTic[] = {0, 0};
int ticLength = 100;

//Time Since Start
unsigned long millisStart = 0;
unsigned int dt = 0;

void setup() {
  Serial.begin(9600);

  //init leds
  FastLED.addLeds<LED_TYPE1, DATA_PIN1, CLK_PIN1, COLOR_ORDER1>(leds1, NUM_LEDS1).setCorrection(TypicalLEDStrip);
  FastLED.addLeds<LED_TYPE2, DATA_PIN2, CLK_PIN2, COLOR_ORDER2>(leds2, NUM_LEDS2).setCorrection(TypicalLEDStrip);
  delay(1000);
  FastLED.setBrightness(BRIGHTNESS);

  Serial.println("Setting Solid Color");
  ChangeColor(CRGB::Red, 1);
  ChangeColor(CRGB::Green, 0);
  FastLED.show();

  delay(1000);
  //FlashColor(CRGB::Blue, 100, 250, 250, 0);
  //    FlashColor(CRGB::Purple, 100, 300, 100, 1);
  Rainbow(1);
  //  Countdown(CRGB::White, 8000, 1);
  Countdown(CRGB::White, 4000, 0);
}

void loop() {
  dt = millis() - millisStart;
  millisStart = millis();
  updateEventManager(dt);
  updateLeds(dt);
  FastLED.show();
  //Serial.println(dt);
}
/*
  --------------------------------------------------------------------------------------------------------------------------
   This function flashes a color on the LEDs
   <color> is the color the light will flash
   <annount> is the ammount of times it will flash
   <timePerFlash> is the time the color is lit
   <timeBetweenFkash> is the time the color is off (if your only flashing once it dose not matter what you set this to)
   <ledStrip> is the stip it will flash on (0 or 1)
  --------------------------------------------------------------------------------------------------------------------------
*/
void FlashColor(CRGB color, int ammount, int _timePerFlash, int _timeBetweenFlash, int ledStrip)
{
  ledColor[ledStrip] = color;
  flashAmmount[ledStrip] = ammount;
  timePerFlash[ledStrip] = _timePerFlash;
  timeBetweenFlash[ledStrip] = _timeBetweenFlash;
  isFlashLit[ledStrip] = false;

  timerCallback[ledStrip] = FLASHCOLOR;
  timer[ledStrip] = 100;

  // Serial.println(timeBetweenFlash[ledStrip]);
}

/*
  --------------------------------------------------------------------------------------------------------------------------
   This function counts down
   <color> is the color the light
   <_time> is the ammount of time it will take
   <ledStrip> is the stip it will count down on
  --------------------------------------------------------------------------------------------------------------------------
*/
void Countdown(CRGB color, int _time, int ledStrip)
{
  ledState[ledStrip] = COUNTING;
  countdownAmmount[ledStrip] = 100;
  PercentagePerTic[ledStrip] = 100 / (_time / ticLength);
  ledColor[ledStrip] = color;
  ChangeColor(color, ledStrip);

  timer[ledStrip] = ticLength;
  timerCallback[ledStrip] = COUNTDOWN;
}

/*
  --------------------------------------------------------------------------------------------------------------------------
   This function makes the leds Rainbow
   <ledStrip> is the stip it will light on (0 or 1)
  --------------------------------------------------------------------------------------------------------------------------
*/
void Rainbow(int ledStrip)
{
  ledState[ledStrip] = RAINBOW;
}

/*
  --------------------------------------------------------------------------------------------------------------------------
   This function changes the color of the LEDs to a solid color
   <color> is the color the light will be
   <ledStrip> is the stip it will color (0 or 1)
  --------------------------------------------------------------------------------------------------------------------------
*/
void ChangeColor(CRGB color, int ledStrip)
{
  //ledState[ledStrip] = SOLID;

  if (ledStrip == 0)
  {
    fill_solid(leds1, NUM_LEDS1, color);
  }
  else
  {
    fill_solid(leds2, NUM_LEDS2, color);
  }

  //  FastLED.show();
}


void updateLeds(int deltaTime)
{
  updateRainbow(deltaTime);
  countdownUpdater(deltaTime);
}

void updateRainbow(int deltaTime)
{
  for (int i = 0; i < (sizeof(timer) / sizeof(int)); i++)
  {
    if (ledState[i] == RAINBOW)
    {
      EVERY_N_MILLISECONDS( 20 ) {
        hue[i]++;
      }
      switch (i)
      {
        case 0:
          fill_rainbow(leds1, NUM_LEDS1, hue[i], 2);
          break;

        case 1:
          fill_rainbow(leds2, NUM_LEDS2, hue[i], 2);
          break;
      }
      //FastLED.show();
    }
  }
}

void countdownUpdater(int deltaTime)
{
  for (int i = 0; i < (sizeof(timer) / sizeof(int)); i++)
  {
    if (ledState[i] == COUNTING)
    {
      EVERY_N_MILLISECONDS( 5 ) {
        switch (i)
        {
          case 0:
            fadeToBlackBy(leds1, NUM_LEDS1, 25);
            break;
          case 1:
            fadeToBlackBy(leds2, NUM_LEDS2, 25);
            break;
        }
      }
      int maxLed = 0;
      switch (i)
      {
        case 0:
          maxLed = NUM_LEDS1 - (((NUM_LEDS1 * 100) * (100 - countdownAmmount[0])) / 10000);
          //Serial.println(maxLed);
          for (int i2 = 0; i2 < maxLed; i2++)
          {
            leds1[i2] = ledColor[i];
          }

          break;
        case 1:
          maxLed = NUM_LEDS2 - (((NUM_LEDS2 * 100) * (100 - countdownAmmount[1])) / 10000);
          Serial.println(maxLed);
          for (int i2 = 0; i2 < maxLed; i2++)
          {
            leds2[i2] = ledColor[i];
          }
          break;
      }
    }
  }
}

//led updates
void flashColorUpdate(int ledStrip)
{
  if (flashAmmount[ledStrip] > 0)
  {
    if (isFlashLit[ledStrip])
    {
      ChangeColor(CRGB::Black, ledStrip);
      isFlashLit[ledStrip] = false;
      flashAmmount[ledStrip]--;
      timer[ledStrip] = timeBetweenFlash[ledStrip];
    }
    else
    {
      ChangeColor(ledColor[ledStrip], ledStrip);
      isFlashLit[ledStrip] = true;
      timer[ledStrip] = timePerFlash[ledStrip];

    }
  }
}

/*
   event manager functions
*/
void updateEventManager(int deltaTime)
{
  for (int i = 0; i < (sizeof(timer) / sizeof(int)); i++)
  {
    if (timer[i] > 0)
    {
      timer[i] -= deltaTime;
    }

    if (timer[i] <= 0 && timer[i] != -9999)
    {

      timer[i] = -9999;
      if (timerCallback[i] == FLASHCOLOR)
      {
        flashColorUpdate(i);
      }
      else if (timerCallback[i] == COUNTDOWN)
      {
        if (countdownAmmount[i] > 0)
        {
          countdownAmmount[i] -= PercentagePerTic[i];
          timer[i] = ticLength;
        }
      }
    }
  }
}
