# Gorm V2 Upgrade Documentation

The following documentation covers what needs to be done to use Gorm V2, which is different to Gorm V1.
And what has changed to enable the upgrade from Gorm V1 to Gorm V2.  

# Ongoing Development

As new development is done, changes are made to existing columns, or new columns are added to the structs that support PhotoPrism.  
These structs are turned into tables in DBMS' that Gorm supports.  At the time of writing PhotoPrism supports SQLite and MariaDB, with requests to support PostgreSQL.
Given the requests to support PostgreSQL the way that the Gorm annotations for the structs are used needed to change.  

## type/size annotation

For all future development the type/size Gorm annotation needs to only use the default types that Gorm supports.  
Do not use a database specific datatype like VARBINARY, VARCHAR, MEDIUMBLOB.  
The following tables give an overview of the database type, to Go type, and the required Gorm annotation.  Not all types are listed.  
If you want the complete set, check the [go-gorm source](https://github.com/go-gorm/) for DataTypeOf for each DBMS.  

### MariaDB translation  
| DBMS Type | Go Type | Gorm annotation |
|----------------------|---------|-----------------------|
| SMALLINT | int | type:int;size:16; |
| MEDIUMINT | int | type:int;size:24; |
| INT | int | type:int;size:32; |
| BIGINT | int | |
| SMALLINT UNSIGNED | uint | type:uint;size:16; |
| MEDIUMINT UNSIGNED | uint | type:uint;size:24; |
| INT UNSIGNED | uint | type:uint;size:32; |
| BIGINT UNSIGNED | uint | |
| FLOAT | float32 | |
| DOUBLE | float64 | |
| VARBINARY(125) | string | type:byte;size:125; |
| VARCHAR(60) | string | size:60; |
| BLOB | as required | type:byte;size:65535; |
| MEDIUMBLOB | as required | type:byte;size:66666; |
| LONGBLOB | as required | type:byte;size:16777216; |
| DATETIME | time.Time | |
| DECIMAL(16,2) | float64 | precision:16;scale:2; |



### SQLite translation  
| DBMS Type | Go Type | Gorm annotation |
|----------------------|---------|-----------------------|
| INTEGER (1) | int | |
| TEXT (2) | string | |
| BLOB (3) | as required | type:byte; |
| REAL (4) | float64 | |
| NUMERIC (5) | time.Time | |
|----------------------|---------|-----------------------|
| SMALLINT (1) | int | type:int;size:16; |
| MEDIUMINT (1) | int | type:int;size:24; |
| INT (1) | int | type:int;size:32; |
| BIGINT (1) | int | |
| SMALLINT UNSIGNED (1) | uint | type:uint;size:16; |
| MEDIUMINT UNSIGNED (1) | uint | type:uint;size:24; |
| INT UNSIGNED (1) | uint | type:uint;size:32; |
| BIGINT UNSIGNED (1) | uint | |
| FLOAT (4) | float32 | |
| DOUBLE (4) | float64 | |
| VARBINARY(125) (2) | string | type:byte;size:125; |
| VARCHAR(60) (2) | string | size:60; |
| BLOB (3) | as required | type:byte;size:65535; |
| MEDIUMBLOB (3) | as required | type:byte;size:66666; |
| LONGBLOB (3) | as required | type:byte;size:16777216; |
| DATETIME (5) | time.Time | |
| DECIMAL(16,2) (5) | float64 | precision:16;scale:2; |


The number in the brackets is "Affinity" which SQLite uses to translate a foreign DBMS type into it's base set of 5 types, at top of table above.  

### PostgreSQL translation

| DBMS Type | Go Type | Gorm annotation |
|----------------------|---------|-----------------------|
| SMALLSERIAL | int | size:16;autoIncrement; |
| SERIAL | int | size:32;autoIncrement; |
| BIGSERIAL | int | autoIncrement; |
| SMALLINT | int | size:16; |
| INTEGER | int | size:32; |
| BIGINT | int |  |
| SMALLSERIAL | uint | size:15;autoIncrement; |
| SERIAL | uint | size:31;autoIncrement; |
| BIGSERIAL | uint | autoIncrement; |
| SMALLINT | uint | size:15; |
| INTEGER | uint | size:31; |
| BIGINT | uint |  |
| NUMERIC(16,2) (5) | float64 | precision:16;scale:2; |
| DECIMAL | float64 | |
| VARCHAR(60) | string | size:60; |
| TEXT | string | |
| TIMESTAMPTZ(4) | time.Time | precision:4; |
| TIMESTAMPTZ | time.Time | |
| BYTEA | Bytes | |
| BYTEA | String | type:byte;size:125; |
| BYTEA | as required | type:byte;size:66666; |

## Foreign Keys

Gorm V2's implementation has introduced foreign keys at the database level.  This will ensure that the data relationship between parent and child records is maintained.  But, it also means that you can't create a child record if the parent is not already committed to the database (or added earlier in the same transaction).  

An example of this is that you can't call the Create function on a Details struct, until the Create function on the Photo struct has already been done.  This is NOT a change to the way that PhotoPrism is already developed.  

It is possible to create an instance of a struct that has child structs (eg. Photo and Detail) by including the content of the child struct in the parent struct.  Gorm will then take care of the creation of both records when photo.Create() is called.  
eg.
```
photo := Photo{
     TakenAt:  time.Date(2020, 11, 11, 9, 7, 18, 0, time.UTC),
     TakenSrc: SrcMeta,
     Details:  &Details {
        Keywords:     "nature, frog",
		Notes:        "notes",
     }
}
```

## Queries

The use of 0 to represent FALSE and 1 to represent TRUE in queries shall no longer be done.  Use TRUE/FALSE as appropriate in queries.  

## Managing tables

Gorm V2 uses the Migrator to provide any changes to table structure.  This replaces DropTableIfExists and CreateTable with Migrator().DropTable and Migrator().CreateTable.  See internal/commands/auth_reset.go for an example.  

## Soft Delete

Gorm V2 has changed the struct to support soft deletion.  It now uses a type gorm.DeletedAt which has a Time time.Time and a Valid Boolean to indicate if a record is deleted.  The structure in the database has not changed.  
Valid = true when a record is soft deleted.  The Time will also be populated.  

# Changes made to support Gorm V2

The follow provides an overview on what changes have been made to PhotoPrism to enable Gorm V2.  

## Breaking Changes

There is only 1 known visible change as a result of the implementation of Gorm V2.  
That is in the PhotoPrism cli, where the output previously returned a DeletedDate the following difference will be visible.  
1. Any command that returns a DeletedDate will not return a column for DeletedAt if the record is not deleted.
2. Any command that returns a DeletedDate will return a gorm.DeletedAt structure if the record is deleted.


## Connection Strings

The connection string for SQLite has been changed, with &_foreign_keys=on being added to ensure that foreign keys are enabled within SQLite like they are on MariaDB.  

## Migrator Changes

The migration has moved from a map to an ordered list to ensure that the migration is done in an order that supports foreign keys, instead of randomly.  
In addition to that, the Truncate function has been updated to execute in foreign key order when removing all records from all tables.  This process also resets the intital auto increment value to one.  
__Newly added tables need to be added to these lists.__


## Structs

The following changes have been made to all Gorm related PhotoPrism structs.  
The definition of a Primary Key has changed from primary_key to primaryKey.  
The definition of auto increment has changed from auto_increment to autoIncrement.  
The definition of a foreign key's source has changed from foreignkey to foreignKey.  
The definition of a foreign key's target field has changed from association_foreignkey to references.  
The definition of a many 2 many relationship has changed from association_jointable_foreignkey to a combination of foreignKey, joinForeignKey, References and joinReferences.  
The definition of associations has been removed.  
The definition of a unique index has changed from unique_index to uniqueIndex.  
The definition of the type SMALLINT has changed from type:SMALLINT to type:int;size:16;
The definition of the type VARBINARY has changed from type:VARBINARY(nn) to type:bytes;size:nn.  
The definition of the type VARCHAR has changed from type:VARCHAR(nn) to size:nn.  
The definition of the field DeletedAt has changed from *time.Time to gorm.DeletedAt.  
The definition of PRELOAD has been removed.  
The use of the gorm definition type:DATETIME has been removed (not required).  

### Album

The column Photos type has changed from PhotoAlbums to []PhotoAlbum.  

### User

The column UserShares type has changed from UserShares to []UserShare.  
The columns UserDetails and UserSettings are no longer automatically preloaded.  

### Cell

The column Place is no longer automatically preloaded.  

### Country

The column CountryPhotoID is no longer a required field.  A migration script has been created to change the number 0 to a NULL in the database.  

### Face

The column EmbeddingJSON has had it's gorm specific type changed from type:MEDIUMBLOB to type:bytes;size:66666.  This is to support PostgreSQL and SQLite which use unsized blob types, whilst the number ensures that MariaDB uses a medium_blob type.  

### Marker

The columns EmbeddingsJSON and LandmarksJSON have had their gorm specific types changed from type:MEDIUMBLOB to type:bytes;size:66666.  This is to support PostgreSQL and SQLite which use unsized blob types, whilst the number ensures that MariaDB uses a medium_blob type.  

### Photo

The columns PhotoLat, PhotoLng and PhotoFNumber have had their gorm specific types removed.  
The columns Details, Camera, Lens, Cell and Place have had their explicit assocations removed.  
The columns Keywords and Albums have had many2many relationships defined.  

### PhotoAlbum

The columns Photo and Album have been removed.  The gorm function SetupJoinTable is used to populate the foreign key into the model because this table is not using the primary keys of Photo and Album.  

### PhotoLabel

The columns Photo and Label have had their Pre Load status removed, and replaced with foreign key definitions.  


### Many to Many joins

The structs Photo and Album are connected via PhotoAlbum by SetupJoinTable.  
The structs Photo and Keyword are connected via PhotoKeyword by SetupJoinTable.  
The structs Label and LabelCategory are connected via Category by SetupJoinTable.  

## Queries

With Gorm V1 the assumption that a 0 = FALSE or 1 = TRUE for boolean values had been made.  All cases of this have been changed to using TRUE/FALSE as appropriate.  
