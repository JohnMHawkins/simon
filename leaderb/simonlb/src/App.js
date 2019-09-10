import React from 'react';
import logo from './logo.svg';
import './App.css';


class AppButton extends React.Component{

  render() {
    return(
      <div className="App-button" onClick={function() {alert('click')}}>
        <h4>{this.props.label}</h4>
      </div>
    )
  }
}



function App() {
  return (
    <div className="App">
      <header className="App-header">
        <img src={logo} className="App-logo" alt="logo" />
        <p>
          Edit <code>src/App.js</code> and save to reload.
        </p>
        <a
          className="App-link"
          href="https://reactjs.org"
          target="_blank"
          rel="noopener noreferrer"
        >
          Learn React
        </a>
        <AppButton label="Test" />
      </header>
    </div>
  );
}

export default App;
