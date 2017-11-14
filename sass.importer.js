var path = require("path");

var prefix = "bootswatch/";

module.exports = function(file) {
  if (file.startsWith(prefix)) {
    file = prefix + "dist/"+ process.env.BOOTSWATCH_THEME + "/" + file.slice(prefix.length);
  }

  return {
    file: file
  }
};
