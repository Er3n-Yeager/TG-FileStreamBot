# MongoDB Setup Instructions

## Quick Fix for Heroku Deployment

The error you're seeing is because Go needs to download the MongoDB driver dependencies. Here's how to fix it:

### Step 1: Update Dependencies

Run these commands in your project directory:

```bash
go get go.mongodb.org/mongo-driver/mongo@v1.13.1
go get go.mongodb.org/mongo-driver/bson@v1.13.1
go get go.mongodb.org/mongo-driver/bson/primitive@v1.13.1
go get go.mongodb.org/mongo-driver/mongo/options@v1.13.1
go get github.com/google/uuid@v1.6.0
go mod tidy
```

### Step 2: Commit and Push

```bash
git add go.mod go.sum
git commit -m "Add MongoDB driver dependencies"
git push heroku main
```

## Alternative: If you don't have Go installed locally

You can create/update the `go.sum` file by adding these entries manually, or push the code and let Heroku's buildpack handle it.

## MongoDB Configuration

### Option 1: MongoDB Atlas (Free Cloud - RECOMMENDED for Heroku)

1. Go to https://www.mongodb.com/cloud/atlas
2. Create a free account
3. Create a free cluster
4. Get your connection string
5. Add to Heroku Config Vars:
   ```
   MONGO_URI=mongodb+srv://username:password@cluster.mongodb.net/?retryWrites=true&w=majority
   MONGO_DB_NAME=filestream
   MONGO_COLLECTION=files
   ```

### Option 2: Heroku Add-on (Paid)

```bash
heroku addons:create mongolab:sandbox
# Or
heroku addons:create mongodb:shared-cluster-1x
```

Then set:
```bash
heroku config:set MONGO_DB_NAME=filestream
heroku config:set MONGO_COLLECTION=files
```

### Option 3: Local MongoDB (For Local Testing)

1. Install MongoDB locally
2. Add to `fsb.env`:
   ```env
   MONGO_URI=mongodb://localhost:27017
   MONGO_DB_NAME=filestream
   MONGO_COLLECTION=files
   ```

## Without MongoDB

If you don't want to use MongoDB right now, the bot will still work! Just don't set the `MONGO_URI` variable. The bot will log a warning and continue without database storage.

## Verifying It Works

When you start the bot, you should see:

```
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
🔌 Initializing MongoDB Connection...
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
[MongoDB] === MONGODB INITIALIZATION STARTED ===
[MongoDB] ✅ Successfully connected to MongoDB!
```

When a user sends a file:
```
💾 File saved to MongoDB msg_id=5 title=Video.mp4
```
