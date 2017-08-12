#include <TimerThree.h>
#include <FastLED.h>
#include "HX711.h"

void setup() {
  // put your setup code here, to run once:

}

void loop() {
  // put your main code here, to run repeatedly:

}

/*
--------------------------------------------------------------------------------------------------------------------------
 * This function flashes a color on the LEDs
 * <color> is the color the light will flash
 * <annount> is the ammount of times it will flash
 * <timePerFlash> is the time the color is lit
 * <timeBetweenFkash> is the time the color is off (if your only flashing once it dose not matter what you set this to)
 * <ledStrip> is the stip it will flash on (0 or 1)
--------------------------------------------------------------------------------------------------------------------------
 */
void FlashColor(CRGB color, int ammount, int timePerFlash, int timeBetweenFlash, int ledStrip)
{

}

/*
--------------------------------------------------------------------------------------------------------------------------
 * This function counts down
 * <color> is the color the light
 * <_time> is the ammount of time it will take
 * <ledStrip> is the stip it will count down on
--------------------------------------------------------------------------------------------------------------------------
 */
void CountDown(CRGB color, int _time, int ledStrip)
{
  
}

/*
--------------------------------------------------------------------------------------------------------------------------
 * This function makes the leds Rainbow
 * <ledStrip> is the stip it will light on (0 or 1)
--------------------------------------------------------------------------------------------------------------------------
 */
void Rainbow(int ledStrip)
{

}

/*
--------------------------------------------------------------------------------------------------------------------------
 * This function changes the color of the LEDs to a solid color
 * <color> is the color the light will be
 * <ledStrip> is the stip it will color (0 or 1)
--------------------------------------------------------------------------------------------------------------------------
 */
void ChangeColor(CRGB color, int ledStrip)
{

}

