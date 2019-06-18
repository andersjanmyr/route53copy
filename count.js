fs = require('fs')

data = fs.readFileSync('dns.json')
data = JSON.parse(data)
console.log(data.ResourceRecordSets.length)
