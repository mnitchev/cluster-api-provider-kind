const fs = require('fs');
const path = require('path');
const Diagram = require('cli-diagram');

const kindClusterProvider = new Diagram()
    .box('Provider package')
    .line()

const provider = new Diagram()
    .box(`\n${kindClusterProvider}\ninfrastructure provider \n (running locally)`);

const workload = new Diagram()
    .box('\n    Kind Cluster \n (workload cluster)\n');

const management = new Diagram()
    .box('\n    Kind Cluster \n (management cluster)\n');

const diagram = new Diagram()
    .box(`${kindClusterProvider}\n\n\n\n  infrastructure provider\n    (running locally)\n\n\n\n`)
    .arrow(['-->:     Manage', '<->:  Run against  '], {size: 5})
    .container( workload + '\n' + management);

const yourMachine = new Diagram()
    .box(`Your Machine\n\n${diagram}`);


console.log(yourMachine.draw());
