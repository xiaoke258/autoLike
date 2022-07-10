package dbs

import (
	"context"
	"database/sql"
	"fmt"
	"math"
	"net"
	"os"
	"time"

	"github.com/go-sql-driver/mysql"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
)

type ViaSSHDialer struct {
	client *ssh.Client
}

func (self *ViaSSHDialer) Dial(addr string) (net.Conn, error) {
	return self.client.Dial("tcp", addr)
}

var DB *sql.DB

func init() {

	sshHost := "39.103.164.136"        // SSH Server Hostname/IP
	sshPort := 22                      // SSH Port
	sshUser := "root"                  // SSH Username
	sshPass := "Maooo123"              // Empty string for no password
	dbUser := "root"                   // DB username
	dbPass := "6a34a4084074"           // DB Password
	dbHost := "39.106.115.24:3306"     // DB Hostname/IP
	dbName := "service-flask-template" // Database name
	//dbPass := "6a34A4084074"                                              // DB Password
	//dbHost := "rm-8vb146o30k71t696a.mysql.zhangbei.rds.aliyuncs.com:3306" // DB Hostname/IP
	//dbName := "service-flask-template"                                    // Database name

	var agentClient agent.Agent
	// Establish a connection to the local ssh-agent
	if conn, err := net.Dial("unix", os.Getenv("SSH_AUTH_SOCK")); err == nil {
		defer conn.Close()

		// Create a new instance of the ssh agent
		agentClient = agent.NewClient(conn)
	}

	// The client configuration with configuration option to use the ssh-agent
	sshConfig := &ssh.ClientConfig{
		User: sshUser,
		Auth: []ssh.AuthMethod{},
		HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error {
			return nil
		},
	}

	// When the agentClient connection succeeded, add them as AuthMethod
	if agentClient != nil {
		sshConfig.Auth = append(sshConfig.Auth, ssh.PublicKeysCallback(agentClient.Signers))
	}
	// When there's a non empty password add the password AuthMethod
	if sshPass != "" {
		sshConfig.Auth = append(sshConfig.Auth, ssh.PasswordCallback(func() (string, error) {
			return sshPass, nil
		}))
	}

	// Connect to the SSH Server
	if sshcon, err := ssh.Dial("tcp", fmt.Sprintf("%s:%d", sshHost, sshPort), sshConfig); err == nil {
		//defer sshcon.Close()

		// Now we register the ViaSSHDialer with the ssh connection as a parameter
		//mysql.RegisterDial("mysql+tcp", (&ViaSSHDialer{sshcon}).Dial)
		mysql.RegisterDialContext("mysql+tcp", func(_ context.Context, addr string) (net.Conn, error) {
			dialer := &ViaSSHDialer{sshcon}
			return dialer.Dial(addr)
		})

		// And now we can use our new driver with the regular mysql connection string tunneled through the SSH connection
		if DB, err = sql.Open("mysql", fmt.Sprintf("%s:%s@mysql+tcp(%s)/%s", dbUser, dbPass, dbHost, dbName)); err == nil {
			DB.SetMaxOpenConns(1024)
			DB.SetMaxIdleConns(100)
			fmt.Printf("Successfully connected to the db\n")

			//if rows, err := db.Query("show tables"); err == nil {
			//	for rows.Next() {
			//		var name string
			//		rows.Scan(&name)
			//		fmt.Printf("Name: %s\n", name)
			//	}
			//	rows.Close()
			//} else {
			//	fmt.Printf("Failure: %s", err.Error())
			//}
			//err := db.Ping()
			//fmt.Println(err)
			//db.Close()

		} else {
			fmt.Printf("Failed to connect to the db: %s\n", err.Error())
		}

	} else {
		fmt.Println(err)
	}
}

type CatDev struct {
	UpdatedTs     string `json:"updated_ts" gorm:"column:updated_ts"`
	TimeCatID     int    `json:"time_cat_id" gorm:"column:time_cat_id"`
	Level         int    `json:"level" gorm:"column:level"`
	UserID        int    `json:"user_id" gorm:"column:user_id"`
	CreatedTs     string `json:"created_ts" gorm:"column:created_ts"`
	HungerPercent int    `json:"hunger_percent" gorm:"column:hunger_percent"`
	Name          string `json:"name" gorm:"column:name"`
	ID            int    `json:"id" gorm:"column:id"`
	Evolution     int    `json:"evolution" gorm:"column:evolution"`
	GrowthValue   int    `json:"growth_value" gorm:"column:growth_value"`
	Hunger        int    `json:"hunger" gorm:"column:hunger"`
}

type CatDevSetup struct {
	ID          uint64    `json:"id" gorm:"column:id"`
	Level       int       `json:"level" gorm:"column:level"`
	AddEachHour int       `json:"add_each_hour" gorm:"column:add_each_hour"`
	SubEachHour int       `json:"sub_each_hour" gorm:"column:sub_each_hour"`
	Experience  int       `json:"experience" gorm:"column:experience"`
	Cost        int       `json:"cost" gorm:"column:cost"`
	CreatedTs   time.Time `json:"created_ts" gorm:"column:created_ts"`
	UpdatedTs   time.Time `json:"updated_ts" gorm:"column:updated_ts"`
}

func (m *CatDevSetup) GetSetup() map[int]map[string]int {
	result := make(map[int]map[string]int)
	if rows, err := DB.Query("select level,add_each_hour,sub_each_hour from cat_dev_setup"); err == nil {
		for rows.Next() {
			var level int
			var add_each_hour int
			var sub_each_hour int
			rows.Scan(&level, &add_each_hour, &sub_each_hour)
			half := int(math.Ceil(float64(add_each_hour) * 0.9))
			twoTenth := int(math.Ceil(float64(add_each_hour) * 0.1))
			row := map[string]int{"add": add_each_hour, "sub": sub_each_hour, "half": half, "two-tenth": twoTenth}
			result[level] = row
		}
	} else {
		fmt.Println(err)
	}
	//DB.Close()
	return result
}

func (cd *CatDev) MultiUpdate() {
	cds := CatDevSetup{}
	setup := cds.GetSetup()
	if rows, err := DB.Query("select level,hunger_percent from cat_dev group by level,hunger_percent"); err == nil {
		for rows.Next() {
			var c CatDev
			var add string
			err = rows.Scan(&c.Level, &c.HungerPercent)
			if err == nil {
				l, ok := setup[c.Level]
				if ok {
					fmt.Println(l)
					fmt.Println(c.HungerPercent)
					if c.HungerPercent == 100 {
						add = "add"
					} else if c.HungerPercent == 50 {
						add = "half"
					} else if c.HungerPercent == 20 {
						add = "two-tenth"
					}
					val, ok := l[add]
					if ok {
						fmt.Println(val)
						if row_counts, err := DB.Exec("update cat_dev set growth_value=growth_value+? where level = ? and hunger_percent = ?", val, c.Level, c.HungerPercent); err == nil {
							fmt.Println(row_counts)
						} else {
							fmt.Println(err)
						}
					} else {
						fmt.Println("'add' not found")
					}
				}
				//DB.Exec()
			} else {
				fmt.Println(err)
			}
		}
	} else {
		fmt.Println(err)
	}
	for k, v := range setup {
		level := k
		sub, ok := v["sub"]
		if ok {
			if rowCounts, err := DB.Exec(`update cat_dev set hunger=if(hunger>= abs(?),hunger+?,0),hunger_percent=if(hunger>=50,100,if(hunger >= 20,50,20)) where level = ?`, sub, sub, level); err == nil {
				fmt.Println(rowCounts)
			} else {
				fmt.Println(err)
			}
		} else {
			fmt.Println("'sub not found")
		}
	}

}

type CatLikeNum struct {
	Id        int    `json:"id" gorm:"column:id"`
	CatId     int    `json:"cat_id" gorm:"column:cat_id"`
	Num       int    `json:"num" gorm:"column:num"`
	CreatedTs string `json:"created_ts" gorm:"column:created_ts"`
	UpdatedTs string `json:"updated_ts" gorm:"column:updated_ts"`
}

type NewTimeRecord struct {
	Id          int    `json:"id"`
	FirstClass  string `json:"first_class"`
	SecondClass string `json:"second_class"`
	RecordJson  string `json:"record_json"`
	UserId      int    `json:"user_id"`
	TimeCatId   int    `json:"time_cat_id"`
	CreatedTs   string `json:"created_ts"`
	UpdatedTs   string `json:"updated_ts"`
	TimeCatName string `json:"time_cat_name"`
	HasImg      int    `json:"has_img"`
	LikeNum     int    `json:"like_num"`
}

type CatMoodUserLikeRelation struct {
	Id        int    `json:"id"`
	MoodId    int    `json:"mood_id"`
	UserId    int    `json:"user_id"`
	CatId     int    `json:"cat_id"`
	CreatedTs string `json:"created_ts"`
	UpdatedTs string `json:"updated_ts"`
}

type TimeCat struct {
	Id           int         `json:"id"`
	IsMine       int         `json:"is_mine"`
	CatId        int         `json:"cat_id"`
	UserId       int         `json:"user_id"`
	Name         string      `json:"name"`
	Sex          string      `json:"sex"`
	Birthday     string      `json:"birthday"`
	Color        string      `json:"color"`
	Breed        string      `json:"breed"`
	ImgPath      string      `json:"img_path"`
	ImgName      string      `json:"img_name"`
	CreatedTs    string      `json:"created_ts"`
	UpdatedTs    string      `json:"updated_ts"`
	IsSterilized interface{} `json:"is_sterilized"`
	Blood        interface{} `json:"blood"`
	Code         interface{} `json:"code"`
	CatType      interface{} `json:"cat_type"`
}

func (r *NewTimeRecord) RootTask() {

}
