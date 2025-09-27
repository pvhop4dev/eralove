// MongoDB initialization script for EraLove
db = db.getSiblingDB('eralove');

// Create collections
db.createCollection('users');
db.createCollection('photos');
db.createCollection('events');
db.createCollection('messages');
db.createCollection('match_requests');

// Create indexes for better performance
db.users.createIndex({ "email": 1 }, { unique: true });
db.users.createIndex({ "created_at": 1 });

db.photos.createIndex({ "user_id": 1 });
db.photos.createIndex({ "created_at": -1 });
db.photos.createIndex({ "tags": 1 });

db.events.createIndex({ "user_id": 1 });
db.events.createIndex({ "date": 1 });
db.events.createIndex({ "created_at": -1 });

db.messages.createIndex({ "sender_id": 1 });
db.messages.createIndex({ "receiver_id": 1 });
db.messages.createIndex({ "created_at": -1 });
db.messages.createIndex({ "sender_id": 1, "receiver_id": 1, "created_at": -1 });

db.match_requests.createIndex({ "sender_id": 1 });
db.match_requests.createIndex({ "receiver_id": 1 });
db.match_requests.createIndex({ "status": 1 });
db.match_requests.createIndex({ "created_at": -1 });

print('Database initialized successfully');
