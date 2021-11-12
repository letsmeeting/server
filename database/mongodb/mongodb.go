package mongodb

import (
    "go.mongodb.org/mongo-driver/mongo"
    "go.mongodb.org/mongo-driver/mongo/options"
    "go.mongodb.org/mongo-driver/mongo/readpref"
    "strconv"
    "time"
)

import (
    "context"
    "github.com/jinuopti/lilpop-server/configure"
    . "github.com/jinuopti/lilpop-server/log"
)

func Connect() error {
    conf := configure.GetConfig()

    uri := "mongodb://"
    // Replace the uri string with your MongoDB deployment's connection string.
    if conf.Db.MongoAuth {
        uri += conf.Db.MongoId + ":" + conf.Db.MongoPwd + "@"
    }
    uri += conf.Db.MongoAddr + ":" + strconv.Itoa(conf.Db.MongoPort)
    Logd("uri: %s", uri)

    clientOptions := options.Client().ApplyURI(uri)
    clientOptions.SetMaxPoolSize(100)
    clientOptions.SetMinPoolSize(10)
    clientOptions.SetMaxConnIdleTime(10 * time.Second)

    ctx, cancel := context.WithTimeout(context.Background(), 10 * time.Second)
    defer cancel()

    client, err := mongo.Connect(ctx, clientOptions)
    if err != nil {
        return err
    }
    defer func() {
        if err = client.Disconnect(ctx); err != nil {
            panic(err)
        }
    }()

    // Ping the primary
    if err := client.Ping(ctx, readpref.Primary()); err != nil {
        return err
    }
    Logd("Successfully connected and pinged.")

    return nil
}
