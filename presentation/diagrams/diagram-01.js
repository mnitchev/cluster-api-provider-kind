const fs = require('fs');
const path = require('path');
const Diagram = require('cli-diagram');

const provider = new Diagram()
    .box('infrastructure provider \n (running locally)');

const workload = new Diagram()
    .box('\n    Kind Cluster \n (workload cluster)\n');

const management = new Diagram()
    .box('\n    Kind Cluster \n (management cluster)\n');

const diagram = new Diagram()
    .box('\n\n\n\ninfrastructure provider\n  (running locally)\n\n\n\n')
    .arrow(['<->: Manage (e.g. call kind cli) ', '<->:         Run against'], {size: 5})
    .container( workload + '\n' + management);

const yourMachine = new Diagram()
    .box(`Your Machine\n\n${diagram}`);


console.log(yourMachine.draw());
