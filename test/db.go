package test

import (
	"fmt"

	"github.com/jinzhu/gorm"
	"github.com/stretchr/testify/assert"
	"github.com/traggo/server/database"
	"github.com/traggo/server/model"
)

// InMemoryDB create a in memory database for testing.
func InMemoryDB(t assert.TestingT) *Database {
	db, err := database.New("sqlite3", "file::memory:?mode=memory&cache=shared")
	assert.Nil(t, err)
	return &Database{DB: db, t: t}
}

// Database wraps the gorm.DB and provides helper methods
type Database struct {
	*gorm.DB
	t assert.TestingT
}

// User creates a user
func (d *Database) User(id int) *UserWithDatabase {
	user := d.NewUser(id, fmt.Sprint("test", id), false)
	return &UserWithDatabase{
		User: user,
		DB:   d.DB,
		t:    d.t,
	}
}

// UserWithDatabase wraps gorm.DB and provides helper methods
type UserWithDatabase struct {
	User model.User
	*gorm.DB
	t assert.TestingT
}

// NewDevice creates a device.
func (d *UserWithDatabase) NewDevice(id int, token string, name string) model.Device {
	device := model.Device{
		ID:        id,
		Token:     token,
		UserID:    d.User.ID,
		Name:      name,
		Type:      model.TypeNoExpiry,
		ActiveAt:  Time("2009-06-30T18:30:00Z"),
		CreatedAt: Time("2009-06-30T18:30:00Z"),
	}
	d.Create(&device)
	return device
}

// NewTagDefinition creates a tag definition.
func (d *UserWithDatabase) NewTagDefinition(key string) model.TagDefinition {
	tagDefinition := model.TagDefinition{
		UserID: d.User.ID,
		Key:    key,
		Type:   model.TypeSingleValue,
	}
	d.Create(&tagDefinition)
	return tagDefinition
}

// AssertHasTagDefinition asserts if a tag definition exists.
func (d *UserWithDatabase) AssertHasTagDefinition(key string, exist bool) *UserWithDatabase {
	existActual := !d.DB.
		Where(&model.TagDefinition{Key: key, UserID: d.User.ID}).
		Find(new(model.TagDefinition)).
		RecordNotFound()
	assert.True(d.t, exist == existActual)
	return d
}

// NewUser creates a user
func (d *Database) NewUser(id int, name string, admin bool) model.User {
	return d.NewUserPass(id, name, []byte{1, 2, 3}, admin)
}

// NewUserPass creates a user
func (d *Database) NewUserPass(id int, name string, pass []byte, admin bool) model.User {
	user := model.User{
		ID:    id,
		Name:  name,
		Pass:  pass,
		Admin: admin,
	}
	d.Create(&user)
	return user
}

// RunningTimeSpan adds a time span without end.
func (d *UserWithDatabase) RunningTimeSpan(from string) *TimeSpanWithDatabase {
	timeSpan := model.TimeSpan{
		UserID:        d.User.ID,
		StartUserTime: ModelTime(from).OmitTimeZone(),
		StartUTC:      ModelTime(from).UTC(),
	}

	d.DB.Create(&timeSpan)

	return &TimeSpanWithDatabase{
		User:     d.User,
		TimeSpan: timeSpan,
		DB:       d.DB,
		t:        d.t,
	}
}

// TimeSpan adds a time span.
func (d *UserWithDatabase) TimeSpan(from string, to string) *TimeSpanWithDatabase {
	wrapper := d.RunningTimeSpan(from)
	userTime := ModelTime(to).OmitTimeZone()
	wrapper.TimeSpan.EndUserTime = &userTime
	utc := ModelTime(to).UTC()
	wrapper.TimeSpan.EndUTC = &utc
	d.DB.Save(&wrapper.TimeSpan)
	return wrapper
}

// TimeSpanWithDatabase wraps gorm.DB and provides helper methods
type TimeSpanWithDatabase struct {
	User     model.User
	TimeSpan model.TimeSpan
	t        assert.TestingT
	*gorm.DB
}

// AssertHasTag asserts if the tag exists or not.
func (d *TimeSpanWithDatabase) AssertHasTag(key, value string, exist bool) *TimeSpanWithDatabase {
	existActual := !d.DB.
		Where(&model.TimeSpanTag{Key: key, StringValue: &value, TimeSpanID: d.TimeSpan.ID}).
		Find(new(model.TimeSpanTag)).
		RecordNotFound()
	assert.True(d.t, exist == existActual)
	return d
}

// AssertExists asserts if the tag exists or not.
func (d *TimeSpanWithDatabase) AssertExists(exist bool) *TimeSpanWithDatabase {
	existActual := !d.DB.
		Where(&model.TimeSpan{ID: d.TimeSpan.ID}).
		Find(new(model.TimeSpan)).
		RecordNotFound()
	assert.True(d.t, exist == existActual)
	return d
}

// Tag adds a tag to the time span.
func (d *TimeSpanWithDatabase) Tag(key string, stringValue string) *TimeSpanWithDatabase {
	tag := model.TimeSpanTag{
		TimeSpanID:  d.TimeSpan.ID,
		Key:         key,
		StringValue: &stringValue,
	}
	d.TimeSpan.Tags = append(d.TimeSpan.Tags, tag)
	d.DB.Save(tag)
	return d
}
