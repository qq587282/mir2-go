package db

import (
	"database/sql"
	"fmt"
	"sync"
	"time"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/mattn/go-sqlite3"
)

type Database interface {
	Open() error
	Close() error
	Ping() error
	
	CreateAccount(loginID, loginPW, userName string) error
	GetAccount(loginID string) (*Account, error)
	UpdateAccount(acc *Account) error
	DeleteAccount(loginID string) error
	
	CreateCharacter(char *Character) error
	GetCharacter(charID int32) (*Character, error)
	GetCharacterByName(name string) (*Character, error)
	GetCharactersByAccount(accountID int32) ([]*Character, error)
	UpdateCharacter(char *Character) error
	DeleteCharacter(charID int32) error
	
	SavePlayerData(charID int32, data []byte) error
	LoadPlayerData(charID int32) ([]byte, error)
	
	SaveHeroData(charID int32, heroName string, data []byte) error
	LoadHeroData(charID int32, heroName string) ([]byte, error)
	
	SaveGlobalVar(name string, value string) error
	LoadGlobalVar(name string) (string, error)
	
	SaveGuild(guild *GuildData) error
	LoadGuild(guildID int32) (*GuildData, error)
	LoadAllGuilds() ([]*GuildData, error)
	
	SaveQuestFlag(charID int32, flags []byte) error
	LoadQuestFlag(charID int32) ([]byte, error)
}

type DBConfig struct {
	Type     string
	Host     string
	Port     int
	User     string
	Password string
	Database string
}

func NewDatabase(cfg DBConfig) (Database, error) {
	switch cfg.Type {
	case "mysql":
		return NewMySQLDB(cfg)
	case "sqlite":
		return NewSQLiteDB(cfg)
	case "memory":
		return NewMemoryDB(), nil
	default:
		return nil, fmt.Errorf("unsupported database type: %s", cfg.Type)
	}
}

type MySQLDB struct {
	db  *sql.DB
	cfg DBConfig
}

type SQLiteDB struct {
	db  *sql.DB
	cfg DBConfig
}

func NewMySQLDB(cfg DBConfig) (*MySQLDB, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.Database)
	
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open mysql: %w", err)
	}
	
	db.SetMaxOpenConns(100)
	db.SetMaxIdleConns(10)
	db.SetConnMaxLifetime(time.Hour)
	
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping mysql: %w", err)
	}
	
	mysqlDB := &MySQLDB{db: db, cfg: cfg}
	if err := mysqlDB.initTables(); err != nil {
		return nil, fmt.Errorf("failed to init tables: %w", err)
	}
	
	return mysqlDB, nil
}

func (db *MySQLDB) initTables() error {
	tables := []string{
		`CREATE TABLE IF NOT EXISTS account (
			account_id INT AUTO_INCREMENT PRIMARY KEY,
			login_id VARCHAR(30) UNIQUE NOT NULL,
			login_pw VARCHAR(32) NOT NULL,
			user_name VARCHAR(50),
			ssno VARCHAR(20),
			phone_num VARCHAR(20),
			quiz VARCHAR(100),
			answer VARCHAR(100),
			email VARCHAR(100),
			join_date DATETIME,
			last_login DATETIME,
			ip_addr VARCHAR(20),
			safe_code VARCHAR(10),
			block_date DATETIME,
			failed_count INT DEFAULT 0,
			admin_level INT DEFAULT 0,
			INDEX idx_login_id (login_id)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,
		
		`CREATE TABLE IF NOT EXISTS character (
			char_id INT AUTO_INCREMENT PRIMARY KEY,
			account_id INT NOT NULL,
			name VARCHAR(30) UNIQUE NOT NULL,
			job TINYINT,
			gender TINYINT,
			level INT,
			gold INT,
			map_name VARCHAR(30),
			x INT,
			y INT,
			hp INT,
			mp INT,
			exp BIGINT,
			hair TINYINT,
			clothes INT,
			weapon INT,
			direction TINYINT,
			pk_point INT,
			delete_time DATETIME,
			create_time DATETIME,
			last_login DATETIME,
			login_status INT DEFAULT 0,
			INDEX idx_account_id (account_id),
			INDEX idx_name (name)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,
		
		`CREATE TABLE IF NOT EXISTS player_data (
			char_id INT PRIMARY KEY,
			data MEDIUMBLOB,
			update_time DATETIME,
			FOREIGN KEY (char_id) REFERENCES character(char_id) ON DELETE CASCADE
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,
		
		`CREATE TABLE IF NOT EXISTS hero_data (
			hero_id INT AUTO_INCREMENT PRIMARY KEY,
			char_id INT NOT NULL,
			hero_name VARCHAR(30) NOT NULL,
			data MEDIUMBLOB,
			update_time DATETIME,
			UNIQUE KEY uk_char_hero (char_id, hero_name),
			FOREIGN KEY (char_id) REFERENCES character(char_id) ON DELETE CASCADE
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,
		
		`CREATE TABLE IF NOT EXISTS global_var (
			var_name VARCHAR(100) PRIMARY KEY,
			var_value TEXT,
			update_time DATETIME
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,
		
		`CREATE TABLE IF NOT EXISTS guild (
			guild_id INT PRIMARY KEY,
			name VARCHAR(50) UNIQUE NOT NULL,
			leader VARCHAR(30),
			level INT DEFAULT 1,
			exp BIGINT DEFAULT 0,
			gold INT DEFAULT 0,
			castle_id INT DEFAULT 0,
			notice TEXT,
			create_time DATETIME,
			INDEX idx_name (name)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,
		
		`CREATE TABLE IF NOT EXISTS guild_member (
			guild_id INT NOT NULL,
			member_name VARCHAR(30) NOT NULL,
			rank INT DEFAULT 4,
			rank_name VARCHAR(50),
			join_date DATETIME,
			contribute INT DEFAULT 0,
			offer INT DEFAULT 0,
			PRIMARY KEY (guild_id, member_name),
			FOREIGN KEY (guild_id) REFERENCES guild(guild_id) ON DELETE CASCADE
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,
		
		`CREATE TABLE IF NOT EXISTS quest_flag (
			char_id INT PRIMARY KEY,
			flags BLOB,
			FOREIGN KEY (char_id) REFERENCES character(char_id) ON DELETE CASCADE
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,
	}
	
	for _, sql := range tables {
		if _, err := db.db.Exec(sql); err != nil {
			return err
		}
	}
	
	return nil
}

func (db *MySQLDB) Open() error { return nil }
func (db *MySQLDB) Close() error { return db.db.Close() }
func (db *MySQLDB) Ping() error { return db.db.Ping() }

func (db *MySQLDB) CreateAccount(loginID, loginPW, userName string) error {
	_, err := db.db.Exec(`
		INSERT INTO account (login_id, login_pw, user_name, join_date)
		VALUES (?, ?, ?, NOW())`,
		loginID, loginPW, userName)
	return err
}

func (db *MySQLDB) GetAccount(loginID string) (*Account, error) {
	row := db.db.QueryRow(`
		SELECT account_id, login_id, login_pw, user_name, ssno, phone_num,
			   quiz, answer, email, join_date, last_login, ip_addr,
			   safe_code, block_date, failed_count, admin_level
		FROM account WHERE login_id = ?`, loginID)
	
	acc := &Account{}
	var joinDate, lastLogin, blockDate sql.NullTime
	
	err := row.Scan(
		&acc.AccountID, &acc.LoginID, &acc.LoginPW, &acc.UserName,
		&acc.SSNo, &acc.PhoneNum, &acc.Quiz, &acc.Answer, &acc.EMail,
		&joinDate, &lastLogin, &acc.IPAddr, &acc.SafeCode, &blockDate,
		&acc.FailedCount, &acc.Admin)
	
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	
	if joinDate.Valid {
		acc.JoinDate = joinDate.Time
	}
	if lastLogin.Valid {
		acc.LastLogin = lastLogin.Time
	}
	if blockDate.Valid {
		acc.BlockDate = blockDate.Time
	}
	
	return acc, nil
}

func (db *MySQLDB) UpdateAccount(acc *Account) error {
	_, err := db.db.Exec(`
		UPDATE account SET 
			user_name = ?, ssno = ?, phone_num = ?, quiz = ?, answer = ?,
			email = ?, last_login = NOW(), ip_addr = ?, safe_code = ?,
			block_date = ?, failed_count = ?, admin_level = ?
		WHERE login_id = ?`,
		acc.UserName, acc.SSNo, acc.PhoneNum, acc.Quiz, acc.Answer,
		acc.EMail, acc.IPAddr, acc.SafeCode, acc.BlockDate,
		acc.FailedCount, acc.Admin, acc.LoginID)
	return err
}

func (db *MySQLDB) DeleteAccount(loginID string) error {
	_, err := db.db.Exec("DELETE FROM account WHERE login_id = ?", loginID)
	return err
}

func (db *MySQLDB) CreateCharacter(char *Character) error {
	result, err := db.db.Exec(`
		INSERT INTO character (account_id, name, job, gender, level, gold,
			map_name, x, y, hp, mp, exp, hair, clothes, weapon, direction,
			pk_point, create_time)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, NOW())`,
		char.AccountID, char.Name, char.Job, char.Gender, char.Level,
		char.Gold, char.MapName, char.X, char.Y, char.HP, char.MP, char.Exp,
		char.Hair, char.Clothes, char.Weapon, char.Direction, char.PKPoint)
	
	if err != nil {
		return err
	}
	
	id, _ := result.LastInsertId()
	char.CharID = int32(id)
	return nil
}

func (db *MySQLDB) GetCharacter(charID int32) (*Character, error) {
	row := db.db.QueryRow(`
		SELECT char_id, account_id, name, job, gender, level, gold,
			map_name, x, y, hp, mp, exp, hair, clothes, weapon,
			direction, pk_point, delete_time, create_time, last_login
		FROM character WHERE char_id = ? AND delete_time IS NULL`, charID)
	
	return db.scanCharacter(row)
}

func (db *MySQLDB) GetCharacterByName(name string) (*Character, error) {
	row := db.db.QueryRow(`
		SELECT char_id, account_id, name, job, gender, level, gold,
			map_name, x, y, hp, mp, exp, hair, clothes, weapon,
			direction, pk_point, delete_time, create_time, last_login
		FROM character WHERE name = ? AND delete_time IS NULL`, name)
	
	return db.scanCharacter(row)
}

func (db *MySQLDB) GetCharactersByAccount(accountID int32) ([]*Character, error) {
	rows, err := db.db.Query(`
		SELECT char_id, account_id, name, job, gender, level, gold,
			map_name, x, y, hp, mp, exp, hair, clothes, weapon,
			direction, pk_point, delete_time, create_time, last_login
		FROM character WHERE account_id = ? AND delete_time IS NULL
		ORDER BY create_time`, accountID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var chars []*Character
	for rows.Next() {
		char, err := db.scanCharacterRows(rows)
		if err != nil {
			return nil, err
		}
		chars = append(chars, char)
	}
	return chars, rows.Err()
}

func (db *MySQLDB) scanCharacter(row *sql.Row) (*Character, error) {
	char := &Character{}
	var deleteTime, createTime, lastLogin sql.NullTime
	
	err := row.Scan(
		&char.CharID, &char.AccountID, &char.Name, &char.Job, &char.Gender,
		&char.Level, &char.Gold, &char.MapName, &char.X, &char.Y,
		&char.HP, &char.MP, &char.Exp, &char.Hair, &char.Clothes,
		&char.Weapon, &char.Direction, &char.PKPoint, &deleteTime,
		&createTime, &lastLogin)
	
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	
	if deleteTime.Valid {
		char.DeleteTime = deleteTime.Time
	}
	if createTime.Valid {
		char.CreateTime = createTime.Time
	}
	if lastLogin.Valid {
		char.LastLogin = lastLogin.Time
	}
	
	return char, nil
}

func (db *MySQLDB) scanCharacterRows(rows *sql.Rows) (*Character, error) {
	char := &Character{}
	var deleteTime, createTime, lastLogin sql.NullTime
	
	err := rows.Scan(
		&char.CharID, &char.AccountID, &char.Name, &char.Job, &char.Gender,
		&char.Level, &char.Gold, &char.MapName, &char.X, &char.Y,
		&char.HP, &char.MP, &char.Exp, &char.Hair, &char.Clothes,
		&char.Weapon, &char.Direction, &char.PKPoint, &deleteTime,
		&createTime, &lastLogin)
	
	if err != nil {
		return nil, err
	}
	
	if deleteTime.Valid {
		char.DeleteTime = deleteTime.Time
	}
	if createTime.Valid {
		char.CreateTime = createTime.Time
	}
	if lastLogin.Valid {
		char.LastLogin = lastLogin.Time
	}
	
	return char, nil
}

func (db *MySQLDB) UpdateCharacter(char *Character) error {
	_, err := db.db.Exec(`
		UPDATE character SET 
			level = ?, gold = ?, map_name = ?, x = ?, y = ?,
			hp = ?, mp = ?, exp = ?, hair = ?, clothes = ?,
			weapon = ?, direction = ?, pk_point = ?, last_login = NOW()
		WHERE char_id = ?`,
		char.Level, char.Gold, char.MapName, char.X, char.Y,
		char.HP, char.MP, char.Exp, char.Hair, char.Clothes,
		char.Weapon, char.Direction, char.PKPoint, char.CharID)
	return err
}

func (db *MySQLDB) DeleteCharacter(charID int32) error {
	_, err := db.db.Exec("UPDATE character SET delete_time = NOW() WHERE char_id = ?", charID)
	return err
}

func (db *MySQLDB) SavePlayerData(charID int32, data []byte) error {
	_, err := db.db.Exec(`
		INSERT INTO player_data (char_id, data, update_time)
		VALUES (?, ?, NOW())
		ON DUPLICATE KEY UPDATE data = VALUES(data), update_time = NOW()`,
		charID, data)
	return err
}

func (db *MySQLDB) LoadPlayerData(charID int32) ([]byte, error) {
	var data []byte
	err := db.db.QueryRow("SELECT data FROM player_data WHERE char_id = ?", charID).Scan(&data)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return data, err
}

func (db *MySQLDB) SaveHeroData(charID int32, heroName string, data []byte) error {
	_, err := db.db.Exec(`
		INSERT INTO hero_data (char_id, hero_name, data, update_time)
		VALUES (?, ?, ?, NOW())
		ON DUPLICATE KEY UPDATE data = VALUES(data), update_time = NOW()`,
		charID, heroName, data)
	return err
}

func (db *MySQLDB) LoadHeroData(charID int32, heroName string) ([]byte, error) {
	var data []byte
	err := db.db.QueryRow("SELECT data FROM hero_data WHERE char_id = ? AND hero_name = ?",
		charID, heroName).Scan(&data)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return data, err
}

func (db *MySQLDB) SaveGlobalVar(name, value string) error {
	_, err := db.db.Exec(`
		INSERT INTO global_var (var_name, var_value, update_time)
		VALUES (?, ?, NOW())
		ON DUPLICATE KEY UPDATE var_value = VALUES(var_value), update_time = NOW()`,
		name, value)
	return err
}

func (db *MySQLDB) LoadGlobalVar(name string) (string, error) {
	var value string
	err := db.db.QueryRow("SELECT var_value FROM global_var WHERE var_name = ?", name).Scan(&value)
	if err == sql.ErrNoRows {
		return "", nil
	}
	return value, err
}

type GuildData struct {
	GuildID   int32
	Name      string
	Leader    string
	Level     int
	Exp       uint64
	Gold      int32
	CastleID  int32
	Notice    string
	CreateTime time.Time
	Members   []*GuildMemberData
}

type GuildMemberData struct {
	Name       string
	Rank       int
	RankName   string
	JoinDate   time.Time
	Contribute int
	Offer      int
}

func (db *MySQLDB) SaveGuild(guild *GuildData) error {
	tx, err := db.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	
	_, err = tx.Exec(`
		INSERT INTO guild (guild_id, name, leader, level, exp, gold, castle_id, notice, create_time)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE 
			leader = VALUES(leader), level = VALUES(level), exp = VALUES(exp),
			gold = VALUES(gold), castle_id = VALUES(castle_id), notice = VALUES(notice)`,
		guild.GuildID, guild.Name, guild.Leader, guild.Level, guild.Exp,
		guild.Gold, guild.CastleID, guild.Notice, guild.CreateTime)
	if err != nil {
		return err
	}
	
	_, err = tx.Exec("DELETE FROM guild_member WHERE guild_id = ?", guild.GuildID)
	if err != nil {
		return err
	}
	
	for _, m := range guild.Members {
		_, err = tx.Exec(`
			INSERT INTO guild_member (guild_id, member_name, rank, rank_name, join_date, contribute, offer)
			VALUES (?, ?, ?, ?, ?, ?, ?)`,
			guild.GuildID, m.Name, m.Rank, m.RankName, m.JoinDate, m.Contribute, m.Offer)
		if err != nil {
			return err
		}
	}
	
	return tx.Commit()
}

func (db *MySQLDB) LoadGuild(guildID int32) (*GuildData, error) {
	row := db.db.QueryRow(`
		SELECT guild_id, name, leader, level, exp, gold, castle_id, notice, create_time
		FROM guild WHERE guild_id = ?`, guildID)
	
	guild := &GuildData{}
	var createTime sql.NullTime
	
	err := row.Scan(&guild.GuildID, &guild.Name, &guild.Leader, &guild.Level,
		&guild.Exp, &guild.Gold, &guild.CastleID, &guild.Notice, &createTime)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	
	if createTime.Valid {
		guild.CreateTime = createTime.Time
	}
	
	rows, err := db.db.Query(`
		SELECT member_name, rank, rank_name, join_date, contribute, offer
		FROM guild_member WHERE guild_id = ?`, guildID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	for rows.Next() {
		m := &GuildMemberData{}
		var joinDate sql.NullTime
		if err := rows.Scan(&m.Name, &m.Rank, &m.RankName, &joinDate, &m.Contribute, &m.Offer); err != nil {
			return nil, err
		}
		if joinDate.Valid {
			m.JoinDate = joinDate.Time
		}
		guild.Members = append(guild.Members, m)
	}
	
	return guild, nil
}

func (db *MySQLDB) LoadAllGuilds() ([]*GuildData, error) {
	rows, err := db.db.Query(`
		SELECT guild_id, name, leader, level, exp, gold, castle_id, notice, create_time
		FROM guild ORDER BY exp DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var guilds []*GuildData
	for rows.Next() {
		guild := &GuildData{}
		var createTime sql.NullTime
		
		err := rows.Scan(&guild.GuildID, &guild.Name, &guild.Leader, &guild.Level,
			&guild.Exp, &guild.Gold, &guild.CastleID, &guild.Notice, &createTime)
		if err != nil {
			return nil, err
		}
		
		if createTime.Valid {
			guild.CreateTime = createTime.Time
		}
		
		guilds = append(guilds, guild)
	}
	
	return guilds, rows.Err()
}

func (db *MySQLDB) SaveQuestFlag(charID int32, flags []byte) error {
	_, err := db.db.Exec(`
		INSERT INTO quest_flag (char_id, flags)
		VALUES (?, ?)
		ON DUPLICATE KEY UPDATE flags = VALUES(flags)`,
		charID, flags)
	return err
}

func (db *MySQLDB) LoadQuestFlag(charID int32) ([]byte, error) {
	var flags []byte
	err := db.db.QueryRow("SELECT flags FROM quest_flag WHERE char_id = ?", charID).Scan(&flags)
	if err == sql.ErrNoRows {
		return make([]byte, 16), nil
	}
	return flags, err
}

func NewSQLiteDB(cfg DBConfig) (*SQLiteDB, error) {
	db, err := sql.Open("sqlite3", cfg.Database)
	if err != nil {
		return nil, fmt.Errorf("failed to open sqlite: %w", err)
	}
	
	db.SetMaxOpenConns(100)
	db.SetMaxIdleConns(10)
	db.SetConnMaxLifetime(time.Hour)
	
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping sqlite: %w", err)
	}
	
	sqliteDB := &SQLiteDB{db: db, cfg: cfg}
	if err := sqliteDB.initTables(); err != nil {
		return nil, fmt.Errorf("failed to init tables: %w", err)
	}
	
	return sqliteDB, nil
}

func (db *SQLiteDB) initTables() error {
	tables := []string{
		`CREATE TABLE IF NOT EXISTS account (
			account_id INTEGER PRIMARY KEY AUTOINCREMENT,
			login_id TEXT UNIQUE NOT NULL,
			login_pw TEXT NOT NULL,
			user_name TEXT,
			ssno TEXT,
			phone_num TEXT,
			quiz TEXT,
			answer TEXT,
			email TEXT,
			join_date DATETIME,
			last_login DATETIME,
			ip_addr TEXT,
			safe_code TEXT,
			block_date DATETIME,
			failed_count INTEGER DEFAULT 0,
			admin_level INTEGER DEFAULT 0
		)`,
		
		`CREATE TABLE IF NOT EXISTS character (
			char_id INTEGER PRIMARY KEY AUTOINCREMENT,
			account_id INTEGER NOT NULL,
			name TEXT UNIQUE NOT NULL,
			job INTEGER,
			gender INTEGER,
			level INTEGER,
			gold INTEGER,
			map_name TEXT,
			x INTEGER,
			y INTEGER,
			hp INTEGER,
			mp INTEGER,
			exp INTEGER,
			hair INTEGER,
			clothes INTEGER,
			weapon INTEGER,
			direction INTEGER,
			pk_point INTEGER,
			delete_time DATETIME,
			create_time DATETIME,
			last_login DATETIME,
			login_status INTEGER DEFAULT 0
		)`,
		
		`CREATE TABLE IF NOT EXISTS player_data (
			char_id INTEGER PRIMARY KEY,
			data BLOB,
			update_time DATETIME
		)`,
		
		`CREATE TABLE IF NOT EXISTS hero_data (
			hero_id INTEGER PRIMARY KEY AUTOINCREMENT,
			char_id INTEGER NOT NULL,
			hero_name TEXT NOT NULL,
			data BLOB,
			update_time DATETIME,
			UNIQUE(char_id, hero_name)
		)`,
		
		`CREATE TABLE IF NOT EXISTS global_var (
			var_name TEXT PRIMARY KEY,
			var_value TEXT,
			update_time DATETIME
		)`,
		
		`CREATE TABLE IF NOT EXISTS guild (
			guild_id INTEGER PRIMARY KEY,
			name TEXT UNIQUE NOT NULL,
			leader TEXT,
			level INTEGER DEFAULT 1,
			exp INTEGER DEFAULT 0,
			gold INTEGER DEFAULT 0,
			castle_id INTEGER DEFAULT 0,
			notice TEXT,
			create_time DATETIME
		)`,
		
		`CREATE TABLE IF NOT EXISTS guild_member (
			guild_id INTEGER NOT NULL,
			member_name TEXT NOT NULL,
			rank INTEGER DEFAULT 4,
			rank_name TEXT,
			join_date DATETIME,
			contribute INTEGER DEFAULT 0,
			offer INTEGER DEFAULT 0,
			PRIMARY KEY (guild_id, member_name)
		)`,
		
		`CREATE TABLE IF NOT EXISTS quest_flag (
			char_id INTEGER PRIMARY KEY,
			flags BLOB
		)`,
	}
	
	for _, sql := range tables {
		if _, err := db.db.Exec(sql); err != nil {
			return err
		}
	}
	
	return nil
}

func (db *SQLiteDB) Open() error { return nil }
func (db *SQLiteDB) Close() error { return db.db.Close() }
func (db *SQLiteDB) Ping() error { return db.db.Ping() }

func (db *SQLiteDB) CreateAccount(loginID, loginPW, userName string) error {
	_, err := db.db.Exec(`
		INSERT INTO account (login_id, login_pw, user_name, join_date)
		VALUES (?, ?, ?, datetime('now'))`,
		loginID, loginPW, userName)
	return err
}

func (db *SQLiteDB) GetAccount(loginID string) (*Account, error) {
	row := db.db.QueryRow(`
		SELECT account_id, login_id, login_pw, user_name, ssno, phone_num,
			   quiz, answer, email, join_date, last_login, ip_addr,
			   safe_code, block_date, failed_count, admin_level
		FROM account WHERE login_id = ?`, loginID)
	
	acc := &Account{}
	var joinDate, lastLogin, blockDate sql.NullString
	
	err := row.Scan(
		&acc.AccountID, &acc.LoginID, &acc.LoginPW, &acc.UserName,
		&acc.SSNo, &acc.PhoneNum, &acc.Quiz, &acc.Answer, &acc.EMail,
		&joinDate, &lastLogin, &acc.IPAddr, &acc.SafeCode, &blockDate,
		&acc.FailedCount, &acc.Admin)
	
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	
	return acc, nil
}

func (db *SQLiteDB) UpdateAccount(acc *Account) error {
	_, err := db.db.Exec(`
		UPDATE account SET 
			user_name = ?, ssno = ?, phone_num = ?, quiz = ?, answer = ?,
			email = ?, last_login = datetime('now'), ip_addr = ?, safe_code = ?,
			block_date = ?, failed_count = ?, admin_level = ?
		WHERE login_id = ?`,
		acc.UserName, acc.SSNo, acc.PhoneNum, acc.Quiz, acc.Answer,
		acc.EMail, acc.IPAddr, acc.SafeCode, acc.BlockDate,
		acc.FailedCount, acc.Admin, acc.LoginID)
	return err
}

func (db *SQLiteDB) DeleteAccount(loginID string) error {
	_, err := db.db.Exec("DELETE FROM account WHERE login_id = ?", loginID)
	return err
}

func (db *SQLiteDB) CreateCharacter(char *Character) error {
	result, err := db.db.Exec(`
		INSERT INTO character (account_id, name, job, gender, level, gold,
			map_name, x, y, hp, mp, exp, hair, clothes, weapon, direction,
			pk_point, create_time)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, datetime('now'))`,
		char.AccountID, char.Name, char.Job, char.Gender, char.Level,
		char.Gold, char.MapName, char.X, char.Y, char.HP, char.MP, char.Exp,
		char.Hair, char.Clothes, char.Weapon, char.Direction, char.PKPoint)
	
	if err != nil {
		return err
	}
	
	id, _ := result.LastInsertId()
	char.CharID = int32(id)
	return nil
}

func (db *SQLiteDB) GetCharacter(charID int32) (*Character, error) {
	row := db.db.QueryRow(`
		SELECT char_id, account_id, name, job, gender, level, gold,
			map_name, x, y, hp, mp, exp, hair, clothes, weapon,
			direction, pk_point, delete_time, create_time, last_login
		FROM character WHERE char_id = ? AND delete_time IS NULL`, charID)
	
	return db.scanCharacter(row)
}

func (db *SQLiteDB) GetCharacterByName(name string) (*Character, error) {
	row := db.db.QueryRow(`
		SELECT char_id, account_id, name, job, gender, level, gold,
			map_name, x, y, hp, mp, exp, hair, clothes, weapon,
			direction, pk_point, delete_time, create_time, last_login
		FROM character WHERE name = ? AND delete_time IS NULL`, name)
	
	return db.scanCharacter(row)
}

func (db *SQLiteDB) GetCharactersByAccount(accountID int32) ([]*Character, error) {
	rows, err := db.db.Query(`
		SELECT char_id, account_id, name, job, gender, level, gold,
			map_name, x, y, hp, mp, exp, hair, clothes, weapon,
			direction, pk_point, delete_time, create_time, last_login
		FROM character WHERE account_id = ? AND delete_time IS NULL
		ORDER BY create_time`, accountID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var chars []*Character
	for rows.Next() {
		char, err := db.scanCharacterRows(rows)
		if err != nil {
			return nil, err
		}
		chars = append(chars, char)
	}
	return chars, rows.Err()
}

func (db *SQLiteDB) scanCharacter(row *sql.Row) (*Character, error) {
	char := &Character{}
	var deleteTime, createTime, lastLogin sql.NullString
	
	err := row.Scan(
		&char.CharID, &char.AccountID, &char.Name, &char.Job, &char.Gender,
		&char.Level, &char.Gold, &char.MapName, &char.X, &char.Y,
		&char.HP, &char.MP, &char.Exp, &char.Hair, &char.Clothes,
		&char.Weapon, &char.Direction, &char.PKPoint, &deleteTime,
		&createTime, &lastLogin)
	
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	
	return char, nil
}

func (db *SQLiteDB) scanCharacterRows(rows *sql.Rows) (*Character, error) {
	char := &Character{}
	var deleteTime, createTime, lastLogin sql.NullString
	
	err := rows.Scan(
		&char.CharID, &char.AccountID, &char.Name, &char.Job, &char.Gender,
		&char.Level, &char.Gold, &char.MapName, &char.X, &char.Y,
		&char.HP, &char.MP, &char.Exp, &char.Hair, &char.Clothes,
		&char.Weapon, &char.Direction, &char.PKPoint, &deleteTime,
		&createTime, &lastLogin)
	
	if err != nil {
		return nil, err
	}
	
	return char, nil
}

func (db *SQLiteDB) UpdateCharacter(char *Character) error {
	_, err := db.db.Exec(`
		UPDATE character SET 
			level = ?, gold = ?, map_name = ?, x = ?, y = ?,
			hp = ?, mp = ?, exp = ?, hair = ?, clothes = ?,
			weapon = ?, direction = ?, pk_point = ?, last_login = datetime('now')
		WHERE char_id = ?`,
		char.Level, char.Gold, char.MapName, char.X, char.Y,
		char.HP, char.MP, char.Exp, char.Hair, char.Clothes,
		char.Weapon, char.Direction, char.PKPoint, char.CharID)
	return err
}

func (db *SQLiteDB) DeleteCharacter(charID int32) error {
	_, err := db.db.Exec("UPDATE character SET delete_time = datetime('now') WHERE char_id = ?", charID)
	return err
}

func (db *SQLiteDB) SavePlayerData(charID int32, data []byte) error {
	_, err := db.db.Exec(`
		INSERT OR REPLACE INTO player_data (char_id, data, update_time)
		VALUES (?, ?, datetime('now'))`,
		charID, data)
	return err
}

func (db *SQLiteDB) LoadPlayerData(charID int32) ([]byte, error) {
	var data []byte
	err := db.db.QueryRow("SELECT data FROM player_data WHERE char_id = ?", charID).Scan(&data)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return data, err
}

func (db *SQLiteDB) SaveHeroData(charID int32, heroName string, data []byte) error {
	_, err := db.db.Exec(`
		INSERT OR REPLACE INTO hero_data (char_id, hero_name, data, update_time)
		VALUES (?, ?, ?, datetime('now'))`,
		charID, heroName, data)
	return err
}

func (db *SQLiteDB) LoadHeroData(charID int32, heroName string) ([]byte, error) {
	var data []byte
	err := db.db.QueryRow("SELECT data FROM hero_data WHERE char_id = ? AND hero_name = ?",
		charID, heroName).Scan(&data)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return data, err
}

func (db *SQLiteDB) SaveGlobalVar(name, value string) error {
	_, err := db.db.Exec(`
		INSERT OR REPLACE INTO global_var (var_name, var_value, update_time)
		VALUES (?, ?, datetime('now'))`,
		name, value)
	return err
}

func (db *SQLiteDB) LoadGlobalVar(name string) (string, error) {
	var value string
	err := db.db.QueryRow("SELECT var_value FROM global_var WHERE var_name = ?", name).Scan(&value)
	if err == sql.ErrNoRows {
		return "", nil
	}
	return value, err
}

func (db *SQLiteDB) SaveGuild(guild *GuildData) error {
	_, err := db.db.Exec(`
		INSERT OR REPLACE INTO guild (guild_id, name, leader, level, exp, gold, castle_id, notice, create_time)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		guild.GuildID, guild.Name, guild.Leader, guild.Level, guild.Exp,
		guild.Gold, guild.CastleID, guild.Notice, guild.CreateTime)
	return err
}

func (db *SQLiteDB) LoadGuild(guildID int32) (*GuildData, error) {
	row := db.db.QueryRow(`
		SELECT guild_id, name, leader, level, exp, gold, castle_id, notice, create_time
		FROM guild WHERE guild_id = ?`, guildID)
	
	guild := &GuildData{}
	var createTime sql.NullString
	
	err := row.Scan(&guild.GuildID, &guild.Name, &guild.Leader, &guild.Level,
		&guild.Exp, &guild.Gold, &guild.CastleID, &guild.Notice, &createTime)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	
	return guild, nil
}

func (db *SQLiteDB) LoadAllGuilds() ([]*GuildData, error) {
	rows, err := db.db.Query(`
		SELECT guild_id, name, leader, level, exp, gold, castle_id, notice, create_time
		FROM guild ORDER BY exp DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var guilds []*GuildData
	for rows.Next() {
		guild := &GuildData{}
		var createTime sql.NullString
		
		err := rows.Scan(&guild.GuildID, &guild.Name, &guild.Leader, &guild.Level,
			&guild.Exp, &guild.Gold, &guild.CastleID, &guild.Notice, &createTime)
		if err != nil {
			return nil, err
		}
		
		guilds = append(guilds, guild)
	}
	
	return guilds, rows.Err()
}

func (db *SQLiteDB) SaveQuestFlag(charID int32, flags []byte) error {
	_, err := db.db.Exec(`
		INSERT OR REPLACE INTO quest_flag (char_id, flags)
		VALUES (?, ?)`,
		charID, flags)
	return err
}

func (db *SQLiteDB) LoadQuestFlag(charID int32) ([]byte, error) {
	var flags []byte
	err := db.db.QueryRow("SELECT flags FROM quest_flag WHERE char_id = ?", charID).Scan(&flags)
	if err == sql.ErrNoRows {
		return make([]byte, 16), nil
	}
	return flags, err
}

type MemoryDB struct {
	accounts     map[string]*Account
	characters   map[int32]*Character
	charByName   map[string]*Character
	playerData   map[int32][]byte
	heroData     map[string][]byte
	globalVars   map[string]string
	guildData    map[int32]*GuildData
	questFlags   map[int32][]byte
	mu           sync.RWMutex
	nextAccountID int32
	nextCharID   int32
}

func NewMemoryDB() *MemoryDB {
	return &MemoryDB{
		accounts:   make(map[string]*Account),
		characters: make(map[int32]*Character),
		charByName: make(map[string]*Character),
		playerData: make(map[int32][]byte),
		heroData:   make(map[string][]byte),
		globalVars: make(map[string]string),
		guildData:  make(map[int32]*GuildData),
		questFlags: make(map[int32][]byte),
	}
}

func (db *MemoryDB) Open() error { return nil }
func (db *MemoryDB) Close() error { return nil }
func (db *MemoryDB) Ping() error { return nil }

func (db *MemoryDB) CreateAccount(loginID, loginPW, userName string) error {
	db.mu.Lock()
	defer db.mu.Unlock()
	
	if _, exists := db.accounts[loginID]; exists {
		return fmt.Errorf("account already exists")
	}
	
	db.nextAccountID++
	acc := &Account{
		AccountID: db.nextAccountID,
		LoginID:   loginID,
		LoginPW:   loginPW,
		UserName:  userName,
	}
	db.accounts[loginID] = acc
	return nil
}

func (db *MemoryDB) GetAccount(loginID string) (*Account, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()
	return db.accounts[loginID], nil
}

func (db *MemoryDB) UpdateAccount(acc *Account) error {
	db.mu.Lock()
	defer db.mu.Unlock()
	db.accounts[acc.LoginID] = acc
	return nil
}

func (db *MemoryDB) DeleteAccount(loginID string) error {
	db.mu.Lock()
	defer db.mu.Unlock()
	delete(db.accounts, loginID)
	return nil
}

func (db *MemoryDB) CreateCharacter(char *Character) error {
	db.mu.Lock()
	defer db.mu.Unlock()
	
	if _, exists := db.charByName[char.Name]; exists {
		return fmt.Errorf("character name exists")
	}
	
	db.nextCharID++
	char.CharID = db.nextCharID
	db.characters[char.CharID] = char
	db.charByName[char.Name] = char
	return nil
}

func (db *MemoryDB) GetCharacter(charID int32) (*Character, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()
	return db.characters[charID], nil
}

func (db *MemoryDB) GetCharacterByName(name string) (*Character, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()
	return db.charByName[name], nil
}

func (db *MemoryDB) GetCharactersByAccount(accountID int32) ([]*Character, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()
	
	var result []*Character
	for _, ch := range db.characters {
		if ch.AccountID == accountID {
			result = append(result, ch)
		}
	}
	return result, nil
}

func (db *MemoryDB) UpdateCharacter(char *Character) error {
	db.mu.Lock()
	defer db.mu.Unlock()
	db.characters[char.CharID] = char
	db.charByName[char.Name] = char
	return nil
}

func (db *MemoryDB) DeleteCharacter(charID int32) error {
	db.mu.Lock()
	defer db.mu.Unlock()
	
	if ch, exists := db.characters[charID]; exists {
		delete(db.charByName, ch.Name)
	}
	delete(db.characters, charID)
	return nil
}

func (db *MemoryDB) SavePlayerData(charID int32, data []byte) error {
	db.mu.Lock()
	defer db.mu.Unlock()
	db.playerData[charID] = data
	return nil
}

func (db *MemoryDB) LoadPlayerData(charID int32) ([]byte, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()
	return db.playerData[charID], nil
}

func (db *MemoryDB) SaveHeroData(charID int32, heroName string, data []byte) error {
	db.mu.Lock()
	defer db.mu.Unlock()
	key := fmt.Sprintf("%d_%s", charID, heroName)
	db.heroData[key] = data
	return nil
}

func (db *MemoryDB) LoadHeroData(charID int32, heroName string) ([]byte, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()
	key := fmt.Sprintf("%d_%s", charID, heroName)
	return db.heroData[key], nil
}

func (db *MemoryDB) SaveGlobalVar(name string, value string) error {
	db.mu.Lock()
	defer db.mu.Unlock()
	db.globalVars[name] = value
	return nil
}

func (db *MemoryDB) LoadGlobalVar(name string) (string, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()
	return db.globalVars[name], nil
}

func (db *MemoryDB) SaveGuild(guild *GuildData) error {
	db.mu.Lock()
	defer db.mu.Unlock()
	db.guildData[guild.GuildID] = guild
	return nil
}

func (db *MemoryDB) LoadGuild(guildID int32) (*GuildData, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()
	return db.guildData[guildID], nil
}

func (db *MemoryDB) LoadAllGuilds() ([]*GuildData, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()
	
	result := make([]*GuildData, 0, len(db.guildData))
	for _, g := range db.guildData {
		result = append(result, g)
	}
	return result, nil
}

func (db *MemoryDB) SaveQuestFlag(charID int32, flags []byte) error {
	db.mu.Lock()
	defer db.mu.Unlock()
	db.questFlags[charID] = flags
	return nil
}

func (db *MemoryDB) LoadQuestFlag(charID int32) ([]byte, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()
	if flags, ok := db.questFlags[charID]; ok {
		return flags, nil
	}
	return make([]byte, 16), nil
}
