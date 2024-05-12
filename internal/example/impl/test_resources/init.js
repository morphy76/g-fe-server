db.createUser({
    user: 'go',
    pwd: 'go',
    roles: [ { role: 'readWrite', db: 'go_db' } ]
});

db.createCollection('examples', {});

db.examples.createIndex({ name: 1 }, { unique: true });

db.examples.insertOne({
    name: 'test',
    age: 101
});
