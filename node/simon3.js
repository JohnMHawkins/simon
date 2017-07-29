var SerialPort = require('serialport');

////////////////

// set to true to run self-test
var autotest = true;


// the ports the arduinos are connected to
const SIMON_CENTER = 0;

const FIRST_BUTTON = 1; // set this to the index of the first arduino we consider a button
const SIMON_RED    = 1;
const SIMON_GREEN  = 2;
const SIMON_BLUE   = 3;
const SIMON_YELLOW = 4;
const LAST_BUTTON  = 4; // set this to the index of the last arduino we consider a button

var ports = {};
var simonPorts = {};

/*
jumpers:

CENTER: 51
BLUEGREEN: 52
REDYELLOW: 53
*/


// this array will hold the latest weights from the buttons
var buttonWeights = [0, 0, 0, 0, 0]

var rgbs = [
  {}, 
  {"r" : 255, "g" : 0, "b" : 0, "rgb" : "255-0-0" },  // SIMON_RED
  {"r" : 0, "g" : 255, "b" : 0, "rgb" : "0-255-0" },  // SIMON_GREEN
  {"r" : 0, "g" : 0, "b" : 255, "rgb" : "0-0-255" },  // SIMON_BLUE
  {"r" : 255, "g" : 255, "b" : 0, "rgb" : "255-255-0" },  // SIMON_YELLOW
]

// data is an array [] where each element will be sent a:b:c
function sendCommand(buttonId, cmd, data) {
  cmdString = "";
  switch ( buttonId) {
    case SIMON_CENTER:
      cmdString = "CENTER:";
      break;
    case SIMON_RED:
      cmdString = "RED:";
      break;
    case SIMON_GREEN:
      cmdString = "GREEN:";
      break;
    case SIMON_BLUE:
      cmdString = "BLUE:";
      break;
    case SIMON_YELLOW:
      cmdString = "YELLOW:";
      break;
  }

  cmdString += cmd
  if ( data ) {
    i = 0;
    while ( i < data.length) {
      cmdString += ":" + data[i].toString();
      i++;
    }

  }

  console.log("SENDING " + buttonId.toString() + "==" + cmdString);
  simonPorts[buttonId].port.write (cmdString + '\n');

}

////////////////////////////////////
// setup functions
//

var inputStrings = {};

function makePort(comName) {
    console.log("makePort for " + comName)
    inputStrings[comName] = "";
    var port = {
        'comName' : comName,
        'port' : new SerialPort( comName, {baudRate : 115200, autoOpen:false}),
        'name':"",
        'onData' : function(data) {
            console.log("onData for " + comName);
            console.log(data);
            console.log("-----");
            inputStrings[comName] += data.toString();
            console.log("inputStrings[comName]:" +inputStrings[comName]);
            if (inputStrings[comName].indexOf('\r\n') > -1 ) {
              cmd = inputStrings[comName].substring(0, inputStrings[comName].indexOf('\r\n'));
              inputStrings[comName] = inputStrings[comName].substring(inputStrings[comName].indexOf('\r\n') + 2);
              console.log("...new inputStrings[comName]:" + inputStrings[comName]);
              receiveSerialData(comName, cmd);
            }
            //receiveSerialData(comName, data.toString());
        }
    }
    return port;

}


////////////////
// tells the ports what functions should handle each event
function setupHandlers(port) {
    console.log ("setting up ports");

    port.port.on('open', showPortOpen);
    port.port.on('data', port.onData);
    port.port.on('close', showPortClose);
    port.port.on('error', showError);

}



////////////////////////////////////
//
// Helper functions
//
//


function openPort (port) {
  port.port.open(function(error) {
      if (error) {
        console.log('failed to open port: ' + error);
      }
  });
}

function showPortOpen(path, options) {
  console.log('port open. Data rate: ' + path + " : " + options);
}

function showPortClose() {
   console.log('port closed.');
}
 
function showError(error) {
   console.log('Serial port error: ' + error);
}



function stripAlphaChars(source) {
  var out = source.replace(/[^0-9]/g, '');

  return out;
}




////////////////
// the ardie sent us data
function receiveSerialData(comName, cmd) {
    console.log("receiveSerialData:" + comName + ":" + cmd);
    
    console.log("... parsed" + cmd);
    if ( gameState == GS_INIT || gameState == GS_WAIT ) {
        if ( cmd.search("NAME:") > -1 ) {
            if ( cmd.search("CENTER") > -1 ) {
                simonPorts[SIMON_CENTER] = ports[comName];
            }
            if ( cmd.search("REDYELLOW") > -1 ) {
                simonPorts[SIMON_RED] = ports[comName];
                simonPorts[SIMON_YELLOW] = ports[comName];
            }
            if ( cmd.search("BLUEGREEN") > -1 ) {
                simonPorts[SIMON_BLUE] = ports[comName];
                simonPorts[SIMON_GREEN] = ports[comName];
            }
        }
        return;
    }

  var ardId = -1;

  cmdparts = cmd.toString().split(":");
  

  console.log("command " + cmdparts[0]);  
   
  if ( cmdparts[0].indexOf("[CENTER]") > -1 ) {
    ardId = SIMON_CENTER;
  } else if ( cmdparts[0].indexOf("[RED]") > -1 ) {
    ardId = SIMON_RED;
  } else if ( cmdparts[0].indexOf("[GREEN]") > -1 ) {
    ardId = SIMON_GREEN;
  } else if ( cmdparts[0].indexOf("[BLUE]") > -1 ) {
    ardId = SIMON_BLUE;
  } else if ( cmdparts[0].indexOf("[YELLOW]") > -1)  {
    ardId = SIMON_YELLOW;
  }

  if ( cmdparts.length > 1 && ardId > 0) {
    if (cmdparts[1] == "BTTN") {
      if (cmdparts.length > 2) {
        weight = parseInt(cmdparts[2])
        console.log("...get weight " + weight.toString() + " for " + ardId.toString());
        buttonWeights[ardId] = weight;
      }
    }
  }
  

}

////////////////
// reads one button
function readButton(btnId) {
  console.log("readbutton of " + btnId.toString());
  sendCommand(btnId, "READ_BUTTONS", null);
  //simonPorts[btnId].port.write ('READ_BUTTONS' + '\n');
}

////////////////
// reads all buttons.  Resets their weights to unread if specified
function readAllButtons(resetWeights) {

  if (resetWeights) {
    // clear everything
    for ( idx = FIRST_BUTTON; idx <= LAST_BUTTON; idx++ ) {
      buttonWeights[idx] = -1;
    }
  }

  console.log("Read Buttons")
  sendCommand(SIMON_RED, "READ_BUTTONS", null);
  sendCommand(SIMON_GREEN, "READ_BUTTONS", null);
  sendCommand(SIMON_BLUE, "READ_BUTTONS", null);
  sendCommand(SIMON_YELLOW, "READ_BUTTONS", null);

}


////////////////
// returns the button with the most weight, or -1 if not all buttons have reported yet
function getPressedButton() {

  var idx = FIRST_BUTTON;
  var maxWeight = 0;
  var maxWeightIdx = idx;
  while (idx <= LAST_BUTTON) {
    if (buttonWeights[idx] == -1) {
      // not reported yet
      return -1;
    }
    if (buttonWeights[idx] >= maxWeight) {
      maxWeight = buttonWeights[idx];
      maxWeightIdx = idx;
    }
    idx = idx + 1;
  }

  console.log("button pushed is..." + maxWeightIdx.toString());

  return maxWeightIdx;

}


///////
// self-test method, will "push" the right button if ok is true, or the wrong one if false
//
function testFakeButtonPress(ok) {

  console.log("player sequence length " + playerSequence.length.toString())
  correctIdx = simonsSequence[playerSequence.length];

  for ( idx = FIRST_BUTTON; idx <= LAST_BUTTON; idx++ ) {
    wgt = 100
    if (idx == correctIdx) {
      if ( ok ) {
        wgt = 1000;
      }
      else {
        wgt = 10;;
      }
    }
    data = ""
    switch ( idx) {
      case SIMON_RED:
        data = "SIMON_RED : ";
        break;
      case SIMON_GREEN:
        data = "SIMON_GREEN : ";
        break;
      case SIMON_BLUE:
        data = "SIMON_BLUE : ";
        break;
      case SIMON_YELLOW:
        data = "SIMON_YELLOW : ";
        break;
    }
    data = data + "BTTN:" + wgt.toString();
    receiveSerialData("", data)
  }


}

////////////////////////////////////
//
// Output logic
//
// These are the methods that control the lights and sound
//

////////////////
// lights up the specified color and triggers the correct sound
function showColor(coloridx, howLong) {
  console.log("ShowColor: " + coloridx.toString() + " for " + howLong.toString());

  // send to raspberryPi
  // TBD

  // trigger audino
  sendCommand(SIMON_CENTER, "FLASHCOLOR", [ rgbs[coloridx]['rgb'], howLong, 1]);
  
  // temporary check in case we don't have all four colors hooked up 
  if (coloridx < simonPorts.length) {
    sendCommand(coloridx, "FLASHCOLOR", [ rgbs[coloridx]['rgb'], howLong, 1]);
  }

}


// showing the sequence Simon has picked for the player

// where we are in simon's sequence
var simonSequenceIdx = 0;
var simonIntervalTimer = null;
var simonms = 0;

////////////////
// shows the next color and sets a timer
function showNextSimonColor() {
  if ( simonSequenceIdx < simonsSequence.length) {
    showColor(simonsSequence[simonSequenceIdx], simonms);
    simonSequenceIdx++;
  }
  else {
    // we're done, start the player's timer or go back to attract mode
    if ( simonIntervalTimer ) {
      clearInterval(simonIntervalTimer);
    }
    if ( gameState == GS_GAME_OVER) {
      // go back to attract mode
      gameState = GS_ATTRACT;
    }
    else {
      startPlayersTimer();
    }
  }
}


////////////////
// playes the entire simon sequence
function showSimonsSequence(gameOver) {
  if ( gameOver ) {
    gameState = GS_GAME_OVER;
  }
  else {
    gameState = GS_SHOWINGSEQ;  // set this so the game loop doesn't re-enter here
  }
  console.log("Showing simon's sequence");
  simonSequenceIdx = 0;
  // how fast do we show simon?
  simonms = 0;
  if ( gameOver ) {
    simonms = 500;
  }
  else {
    if (simonsSequence.length > 10 ) {
      simonms = 125;
    }
    else if (simonsSequence.length > 6 ) {
      simonms = 250;
    }
    else {
      simonms = 500;
    }
  }
  simonIntervalTimer = setInterval (showNextSimonColor,simonms);
}


////////////////
// starts the timer for the player to pick their button
function startPlayersTimer() {
  gameState = GS_TIMER;
  console.log("Starting player's timer");
  l = simonsSequence.lenth;

  timerms = 3000; // 3 seconds unless...
  if ( l > 15) {
    timerms = 1500; // gotta be quick!
  }

  // send message to center arduino
  // TBD
  sendCommand(SIMON_CENTER, "GS_TIMER", ["255-255-255",  timerms]);
  //simonPorts[SIMON_CENTER].port.write("GS_TIMER:" + timerms.toString());

  // when time is up, read the buttons
  setTimeout(function(){

    readAllButtons(true);
    // uncomment to run self-test sequence
    if ( autotest ) {
      //testFakeButtonPress(playerSequence.length < 15);
    }
  
    gameState = GS_PLAYER;
  }, timerms);

}


////////////////////////////////////
//
// Game logic
//

// the sequence the computer has selected
var simonsSequence = []

// the sequence so far that the player has input
var playerSequence = []


////////////////
// pick a new color to add, then show the sequence
function addNewColor() {
  console.log("Add new color");
  var coloridx = Math.floor(Math.random() * (LAST_BUTTON + 1 - FIRST_BUTTON) + FIRST_BUTTON);
  simonsSequence.push(coloridx);
  console.log(simonsSequence)
  showSimonsSequence(false);

}


////////////////
////////////////
// checks if the player's pattern so far matches
function checkPattern() {
  // note that player sequence will always be less than simon's sequence
  for ( idx = 0; idx < playerSequence.length; idx++  ) {
    if ( playerSequence[idx] != simonsSequence[idx]) {
      return false;
    } 
  }
  return true;

}


// checks which button is pressed and uses that as the player's choice
function makePlayersChoice() {
  var selected = getPressedButton();
  if ( selected > -1 ) {
    gameState = GS_EVALUATING;  // set this so the loop doesn't re-enter this function
    console.log("Make player choice");
    playerSequence.push(selected);

    if ( checkPattern() ) {
      // success, show pushed color
      showColor(selected, 500);
      if ( playerSequence.length == simonsSequence.length ) {
        // they got it!   computer's turn
        console.log("++++ player got entire sequence");
        playerSequence = [];
        gameState = GS_COMPUTER;  // computer's turn
      }
      else {
        // so far so good, but they haven't finished the sequence yet, start timer for next button
        startPlayersTimer();
      }
    

    }
    else {
      // bzzzzzz - failed, game over
      console.log("GAME OVER!!!!");
      gameOver();

    }

  }

  
}

////////////////
// show the game over sequence
function gameOver() {
  // buzz 
  showSimonsSequence(true);

}


////////////////
// start a brand new game
function newGame() {
  console.log("New Game!");
  simonsSequence = [];
  playerSequence = [];
  gameState = GS_COMPUTER;
}






////////////////////////////////////
//
// Main functions and loop
//
//

////////////////
// The first thing we do is find and connect all the arduinos connected as USB/serial ports 
SerialPort.list(function (err, portList) {
 portList.forEach(function( portInfo) {
    console.log( portInfo.comName);
    console.log( portInfo.pnpId);
    console.log( portInfo.manufacturer);
    var newPort = makePort(portInfo.comName);
    ports[portInfo.comName] = newPort;
    
   setupHandlers(newPort);
   openPort (newPort);
    
  });
});


function setupArduinos() {
    for ( i in ports ) {
        var p = ports[i];
        console.log("setting up " + i + " : " + p.toString() );
        p.port.write("NAME:" + '\n');
    }

}

/*
 * Main Simon Game State Loop
 */

// GAME STATES
const GS_INIT       = -1; // waiting to set up
const GS_WAIT       = 0;
const GS_ATTRACT    = 1;  // running attract mode, waiting for RED to be pressed to start the game
const GS_COMPUTER   = 2;  // Player got sequence right, Computer is thinking, adding a new color to the sequence
const GS_SHOWINGSEQ = 3;  // we are showing the sequence.  Just wait...
const GS_TIMER      = 4;  // timer is running down, just wait...
const GS_PLAYER     = 5;  // timer ran our, figure out what button the player pushed
const GS_EVALUATING = 6;  // evaluate if the player pushed the right button
const GS_GAME_OVER  = 7; // player got the sequence wrong...

var gameState = GS_INIT;



////////////////
// this will start the loop 5 seconds after we start up (to allow ports to be initialized), then
// start a 50 ms loop
setTimeout(function(){
  console.log("Starting Loop")
  setInterval (loop,500);
}, 5000);


////////////////
// the main game loop
function loop () {

  switch( gameState) {
    case GS_INIT:
        console.log("gs_init");
        setupArduinos();
        gameState = GS_WAIT;
        break;

        case GS_WAIT:
            bGotAll = true;
            var i = 0;
            while ( i <= LAST_BUTTON ) {
                console.log("check " + i);
                if (simonPorts[i] == null) {
                    bGotAll = false;
                    console.log("dont got all " + i.toString());
                }
                i++;
            }
            console.log("GS_WAIT gotAll = " + bGotAll.toString());
            if ( bGotAll ) {
                gameState = GS_ATTRACT;
            }
            else {
                setupArduinos();
            }
            break;

    case GS_ATTRACT:
      console.log("GS_ATTRACT waiting for red button");
      readButton(SIMON_RED);
      //readButton(SIMON_GREEN);
      if (buttonWeights[SIMON_RED] > 0) {
        console.log ("Button Pushed :: Start Game");
        sendCommand(SIMON_CENTER, "GS_COMPUTER", null);
        //simonPorts[SIMON_CENTER].port.write ('GS_COMPUTER' + '\n');
        newGame();
      }
      ////////////
      if ( autotest ) {
        newGame();
        //setTimeout(newGame, 5000);
      }
      //
      ///////////
      break;

    case GS_COMPUTER:
      // simon is taking his turn. 
      addNewColor();
      break;

    case GS_SHOWINGSEQ:
      // be patient while simon shows the player the sequence
      break;

    case GS_TIMER:
      // waiting for the timer to run down
      break;

    case GS_PLAYER:
      makePlayersChoice();
      break;

    case GS_EVALUATING:
      // checking to see if the player got it right...
      break;

    case GS_GAME_OVER:
      // showing the game over sequence, then returning to attract mode
      break;
  }


}
  