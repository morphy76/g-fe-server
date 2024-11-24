db.createUser({
    user: 'fe_user',
    pwd: 'fe_password',
    roles: [ { role: 'readWrite', db: 'fe_db' } ]
});
