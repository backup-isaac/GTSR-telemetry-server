"use strict"
var MarkovChain = require('markovchain')
  , fs = require('fs')
  , quotes = new MarkovChain(fs.readFileSync('test.txt', 'utf8'))

const express = require('express');
const bodyParser = require('body-parser');
const app = express();
app.use(bodyParser.json());
app.use(bodyParser.urlencoded({ extended: true }));
const server = app.listen(3001, () => {
  console.log('Express server listening on port %d in %s mode',
  server.address().port,
  app.settings.env);
});

var pickWord = function(wordList) {
  var tmpList = Object.keys(wordList);
  var canList = ["maybe", "consider", "probably", "try", "I'm", "kill", "destroy", "should", "fuck", "always", "never", "don't", "think", "no", "yes"]
  if (Math.random() < 0.15) {
    return canList[Math.floor((Math.random()*canList.length))]
  }
  return tmpList[Math.floor((Math.random()*tmpList.length))]
}

app.post('/', (req, res) => {
  let text = req.body.text;
  var result = quotes.start(pickWord).end(25).process();
  let data = {
    response_type: 'in_channel', // public to the channel
    text: result,
  };
  res.json(data);
});
