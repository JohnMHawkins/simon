var SerialPort = require('serialport');

////////////////

// set to true to run self-test
var autotest = false;

var loglevel = 1;

const LOG_ERROR = 0;
const LOG_DEBUG = 1;
const LOG_INFO  = 2;

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

function LOG(level, msg) {
  if (level <= loglevel ) {
    console.log(msg + " (" + new Date().getTime().toString() + ")");
  }

}

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

  LOG(LOG_INFO,"SENDING " + buttonId.toString() + "==" + cmdString);
  simonPorts[buttonId].port.write (cmdString + '\n');

}

function sendCommandToAll(cmd, data) {
  sendCommand(SIMON_CENTER, cmd, data);
  sendCommand(SIMON_RED, cmd, data);
  sendCommand(SIMON_GREEN, cmd, data);
  sendCommand(SIMON_BLUE, cmd, data);
  sendCommand(SIMON_YELLOW, cmd, data);
}

////////////////////////////////////
// setup functions
//

var inputStrings = {};

function makePort(comName) {
    LOG(LOG_DEBUG, "makePort for " + comName)
    inputStrings[comName] = "";
    var port = {
        'comName' : comName,
        'port' : new SerialPort( comName, {baudRate : 9600, autoOpen:false}),
        'onData' : function(data) {
            LOG(LOG_INFO, "onData for " + comName);
            LOG(LOG_INFO, data);
            LOG(LOG_INFO, "-----");
            inputStrings[comName] += data.toString();
            if (inputStrings[comName].indexOf('\r\n') > -1 ) {
              cmd = inputStrings[comName].substring(0, inputStrings[comName].indexOf('\r\n'));
              inputStrings[comName] = inputStrings[comName].substring(inputStrings[comName].indexOf('\r\n') + 2);
              receiveSerialData(comName, cmd);
            }
            
        }
    }
    return port;

}


////////////////
// tells the ports what functions should handle each event
function setupHandlers(port) {
    LOG(LOG_DEBUG, "setting up ports");

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
        LOG(LOG_ERROR, 'failed to open port: ' + error);
      }
  });
}

function showPortOpen(path, options) {
  LOG(LOG_ERROR, 'port open. Data rate: ' + path + " : " + options);
}

function showPortClose() {
   LOG(LOG_ERROR, 'port closed.');
}
 
function showError(error) {
   LOG(LOG_ERROR, 'Serial port error: ' + error);
}



function stripAlphaChars(source) {
  var out = source.replace(/[^0-9]/g, '');

  return out;
}




////////////////
// the ardie sent us data
function receiveSerialData(comName, cmd) {
    LOG(LOG_INFO, "receiveSerialData:" + comName + ":" + cmd);
    
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

  //console.log("command " + cmdparts[0] );  
   
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
        LOG(LOG_INFO, "...get weight " + weight.toString() + " for " + ardId.toString());
        buttonWeights[ardId] = weight;
      }
    }
  }
  

}

////////////////
// reads one button
function readButton(btnId) {
  LOG(LOG_INFO, "readbutton of " + btnId.toString());
  sendCommand(btnId, "READ_BUTTONS", null);
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

  LOG(LOG_INFO, "Read Buttons")
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

  LOG(LOG_DEBUG, "button pushed is..." + maxWeightIdx.toString());

  return maxWeightIdx;

}


///////
// self-test method, will "push" the right button if ok is true, or the wrong one if false
//
function testFakeButtonPress(ok) {

  LOG(LOG_DEBUG, "player sequence length " + playerSequence.length.toString())
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
  LOG(LOG_DEBUG, "ShowColor: " + coloridx.toString() + " for " + howLong.toString());

  // send to raspberryPi
  // TBD

  // trigger arduino
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
    if ( gameState == GS_REPLAY) {
      // go back to attract mode
      gotoAttract();
      
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
    gameState = GS_REPLAY;
  }
  else {
    gameState = GS_SHOWINGSEQ;  // set this so the game loop doesn't re-enter here
  }
  LOG(LOG_DEBUG, "Showing simon's sequence");
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
  LOG(LOG_DEBUG, "Starting player's timer");
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
  LOG(LOG_DEBUG, "Add new color");
  var coloridx = Math.floor(Math.random() * (LAST_BUTTON + 1 - FIRST_BUTTON) + FIRST_BUTTON);
  simonsSequence.push(coloridx);
  LOG(LOG_DEBUG, simonsSequence)
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
    LOG(LOG_DEBUG, "Make player choice");
    playerSequence.push(selected);

    if ( checkPattern() ) {
      // success, show pushed color
      showColor(selected, 500);
      if ( playerSequence.length == simonsSequence.length ) {
        // they got it!   computer's turn
        LOG(LOG_DEBUG, "++++ player got entire sequence");
        playerSequence = [];
        setTimeout(function () {
          gameState = GS_COMPUTER;  // computer's turn
        }, 600);
        
      }
      else {
        // so far so good, but they haven't finished the sequence yet, start timer for next button
        startPlayersTimer();
      }
    

    }
    else {
      // bzzzzzz - failed, game over
      LOG(LOG_DEBUG, "PLAYER FAILED!!!!!!");
      gameOver();

    }

  }

  
}

////////////////
// show the game over sequence
function gameOver() {
  // buzz 
  // send GS_GAME_OVER to all arduinos
  LOG(LOG_DEBUG, "calling gameOver, gamestate = " + gameState.toString())
  sendCommandToAll("FLASHCOLOR", [rgbs[SIMON_RED]['rgb'],500,3] );

  setTimeout(function(){
    LOG(LOG_DEBUG, "Setting GS to GS_REPLAY");
    gameState = GS_REPLAY;
    showSimonsSequence(true);

  }, 3000);

  
}

function gotoAttract() {
  LOG(LOG_DEBUG, "gotoAttract");
  sendCommandToAll("GS_ATTRACT", []);
  gameState = GS_ATTRACT;
}

////////////////
// start a brand new game
function newGame() {
  LOG(LOG_DEBUG, "New Game!");
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
    LOG(LOG_DEBUG, portInfo.comName);
    LOG(LOG_DEBUG, portInfo.pnpId);
    LOG(LOG_DEBUG, portInfo.manufacturer);
    var newPort = makePort(portInfo.comName);
    ports[portInfo.comName] = newPort;
    
   setupHandlers(newPort);
   openPort (newPort);
    
  });
});


function setupArduinos() {
    for ( i in ports ) {
        var p = ports[i];
        LOG(LOG_DEBUG, "setting up " + i + " : " + p.toString() );
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
const GS_REPLAY     = 8; // replaying the final sequence

var gameState = GS_INIT;



////////////////
// this will start the loop 5 seconds after we start up (to allow ports to be initialized), then
// start a 50 ms loop
setTimeout(function(){
  LOG(LOG_DEBUG, "Starting Loop")
  setInterval (loop,500);
}, 5000);


////////////////
// the main game loop
function loop () {

  switch( gameState) {
    case GS_INIT:
        LOG(LOG_DEBUG, "gs_init");
        setupArduinos();
        gameState = GS_WAIT;
        break;

        case GS_WAIT:
            bGotAll = true;
            var i = 0;
            while ( i <= LAST_BUTTON ) {
                LOG(LOG_DEBUG, "check " + i);
                if (simonPorts[i] == null) {
                    bGotAll = false;
                    LOG(LOG_DEBUG, "dont got all " + i.toString());
                }
                i++;
            }
            LOG(LOG_DEBUG, "GS_WAIT gotAll = " + bGotAll.toString());
            if ( bGotAll ) {
              gotoAttract();
            }
            else {
                setupArduinos();
            }
            break;

    case GS_ATTRACT:
      LOG(LOG_INFO, "GS_ATTRACT waiting for red button");
      readButton(SIMON_RED);
      //readButton(SIMON_GREEN);
      if (buttonWeights[SIMON_RED] > 0) {
        LOG(LOG_DEBUG, "Button Pushed :: Start Game");
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

    case GS_REPLAY:
      break;
  }


}
  