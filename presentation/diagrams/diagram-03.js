const fs = require('fs');
const path = require('path');
const Diagram = require('cli-diagram');

const kindClusterProvider = new Diagram()
    .box('Provider package')
    .line();

const provider = new Diagram()
    .box(`KindCluster Controller\n${kindClusterProvider}\n`)
    .line();

const workload = new Diagram()
    .box('\n\n\n\n\n    Workload Kind Cluster(s)   \n\n\n\n\n\n');

const management = new Diagram()
    .box(`Management kind cluster\n${provider}`);

const diagram = new Diagram()
    .container(management)
    .arrow(['-->:Manage'], {size: 5})
    .container( workload );


console.log(diagram.draw());
