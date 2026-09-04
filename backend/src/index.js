const express = require('express');
const app = express();
const port = 7273;

app.get('/', (req, res) => {
  res.send('hi');
});

app.listen(port, () => {
  console.log(`Disco MUD backend listening on port ${port}`);
});
