// Validate nash's grammar

var fs = require('fs');

function getExamples(rootDir, cb) {
    fs.readdir(rootDir, function(err, files) { 
        var scripts = []; 
        for (var i = 0; i < files.length; i++) { 
            var file = files[i]; 
            if (file.endsWith('.sh')) { 
                scripts.push(rootDir + '/' + file);                
            }
        }
        cb(scripts);
    });
}

var ohm = require('ohm-js');
var contents = fs.readFileSync('nash.ohm');
var nashGrammar = ohm.grammar(contents);

// test the grammar of each example
getExamples(".", function(files) {
    for (var i = 0; i < files.length; i++) {
        var file = files[i];
        var scriptSource = fs.readFileSync(file);
        var m = nashGrammar.match(scriptSource);
        if (m.succeeded()) {
            console.log(file, ": ok");
        } else {
            console.error(file, ": fail");
            //console.error(nashGrammar.trace(scriptSource).toString());
        }
    }
});
