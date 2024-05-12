db.createUser({
    user: 'go',
    pwd: 'go',
    roles: [ { role: 'readWrite', db: 'go_db' } ]
});
