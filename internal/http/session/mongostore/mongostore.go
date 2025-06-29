package mongostore

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/gorilla/securecookie"
	"github.com/gorilla/sessions"
	"github.com/morphy76/g-fe-server/internal/auth"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

var (
	// ErrInvalidID is returned when an invalid session ID is encountered.
	ErrInvalidID = errors.New("mongostore: invalid session id")
)

// Session represents a session stored in MongoDB.
type Session struct {
	// ID is the unique identifier for the session, stored as an ObjectID.
	ID bson.ObjectID `bson:"_id,omitempty"`
	// Data is the encoded session data.
	Data string
	// Modified is the timestamp when the session was last modified.
	Modified time.Time
	// IAMIssuer is the issuer of the session, typically an OIDC issuer.
	IAMIssuer string `bson:"iam_issuer,omitempty"`
	// IAMSubject is the subject of the session, typically a user ID.
	IAMSubject string `bson:"iam_subject,omitempty"`
	// IAMSessionID is the session ID from the OIDC provider.
	IAMSessionID string `bson:"iam_session_id,omitempty"`
}

// MongoStore is a session store that uses MongoDB to store session data.
type MongoStore struct {
	Codecs  []securecookie.Codec
	Options *sessions.Options
	Token   TokenGetSeter
	coll    *mongo.Collection
}

// NewMongoStore creates a new MongoStore instance with the provided MongoDB collection and session options.
func NewMongoStore(
	c *mongo.Collection,
	sessionOptions *sessions.Options,
	keyPairs ...[]byte,
) *MongoStore {
	store := &MongoStore{
		Codecs:  securecookie.CodecsFromPairs(keyPairs...),
		Options: sessionOptions,
		Token:   &CookieToken{},
		coll:    c,
	}

	store.MaxAge(sessionOptions.MaxAge)
	for _, codec := range store.Codecs {
		asSecrureCookie, ok := codec.(*securecookie.SecureCookie)
		if ok {
			asSecrureCookie.MaxLength(int(^uint(0) >> 1))
		}
	}

	go func() {
		// TODO: this operation should be executed by just the leader of the replicaset
		ttlInSeconds := evalTTL(sessionOptions.MaxAge)
		expireAfter := int32(ttlInSeconds.Seconds())

		expirationIndexModel := mongo.IndexModel{
			Keys: bson.M{"modified": 1},
			Options: options.Index().
				SetExpireAfterSeconds(expireAfter).
				SetName("session_expire_index"),
		}

		c.Indexes().DropOne(context.Background(), "session_expire_index")
		c.Indexes().CreateOne(context.Background(), expirationIndexModel)
	}()

	return store
}

func evalTTL(maxAge int) time.Duration {
	if maxAge <= 0 {
		return 24 * time.Hour
	}
	return time.Duration(maxAge) * time.Second
}

// Get retrieves a session by name from the request.
func (m *MongoStore) Get(r *http.Request, name string) (
	*sessions.Session, error) {
	return sessions.GetRegistry(r).Get(m, name)
}

// New creates a new session with the given name. If a session already exists
func (m *MongoStore) New(r *http.Request, name string) (
	*sessions.Session, error) {
	session := sessions.NewSession(m, name)
	session.Options = &sessions.Options{
		Path:        m.Options.Path,
		MaxAge:      m.Options.MaxAge,
		Domain:      m.Options.Domain,
		Secure:      m.Options.Secure,
		HttpOnly:    m.Options.HttpOnly,
		Partitioned: m.Options.Partitioned,
		SameSite:    m.Options.SameSite,
	}
	session.IsNew = true
	var err error
	if cook, errToken := m.Token.GetToken(r, name); errToken == nil {
		err = securecookie.DecodeMulti(name, cook, &session.ID, m.Codecs...)
		if err == nil {
			err = m.load(session)
			if err == nil {
				session.IsNew = false
			} else {
				err = nil
			}
		}
	}
	return session, err
}

// Save saves the session to the MongoDB collection and sets the session cookie in the response.
func (m *MongoStore) Save(r *http.Request, w http.ResponseWriter,
	session *sessions.Session) error {
	if session.Options.MaxAge < 0 {
		if err := m.delete(session); err != nil {
			return err
		}
		m.Token.SetToken(w, session.Name(), "", session.Options)
		return nil
	}

	if session.ID == "" {
		session.ID = bson.NewObjectID().Hex()
	}

	if err := m.upsert(session); err != nil {
		return err
	}

	encoded, err := securecookie.EncodeMulti(session.Name(), session.ID,
		m.Codecs...)
	if err != nil {
		return err
	}

	m.Token.SetToken(w, session.Name(), encoded, session.Options)
	return nil
}

// MaxAge sets the maximum age for the session cookies and updates the codecs accordingly.
func (m *MongoStore) MaxAge(age int) {
	m.Options.MaxAge = age

	for _, codec := range m.Codecs {
		if sc, ok := codec.(*securecookie.SecureCookie); ok {
			sc.MaxAge(age)
		}
	}
}

func (m *MongoStore) load(session *sessions.Session) error {

	objID, err := bson.ObjectIDFromHex(session.ID)
	if err != nil {
		return err
	}

	s := Session{}
	if err := m.coll.FindOne(context.Background(), bson.M{"_id": objID}).Decode(&s); err != nil {
		return err
	}

	if err := securecookie.DecodeMulti(session.Name(), s.Data, &session.Values,
		m.Codecs...); err != nil {
		return err
	}

	return nil
}

func (m *MongoStore) upsert(session *sessions.Session) error {

	objID, err := bson.ObjectIDFromHex(session.ID)
	if err != nil {
		return err
	}

	var modified time.Time
	if val, ok := session.Values["modified"]; ok {
		modified, ok = val.(time.Time)
		if !ok {
			return errors.New("mongostore: invalid modified value")
		}
	} else {
		modified = time.Now()
	}

	encoded, err := securecookie.EncodeMulti(session.Name(), session.Values, m.Codecs...)
	if err != nil {
		return err
	}

	s := Session{
		ID:           objID,
		Data:         encoded,
		Modified:     modified,
		IAMIssuer:    session.Values[auth.SessionKeyIssuer].(string),
		IAMSubject:   session.Values[auth.SessionKeySubject].(string),
		IAMSessionID: session.Values[auth.SessionKeySessionID].(string),
	}

	opts := options.UpdateOne().SetUpsert(true)
	filter := bson.M{"_id": s.ID}
	updateData := bson.M{"$set": s}

	if _, err = m.coll.UpdateOne(context.Background(), filter, updateData, opts); err != nil {
		return err
	}

	return nil
}

func (m *MongoStore) delete(session *sessions.Session) error {

	objID, err := bson.ObjectIDFromHex(session.ID)
	if err != nil {
		return err
	}

	_, err = m.coll.DeleteOne(context.Background(), bson.M{"_id": objID})
	return err
}
