package handleers

import (
	"StudentService/database"
	"StudentService/models"
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"golang.org/x/sync/singleflight"
	"gorm.io/gorm"
)

var (
	studentGroup  singleflight.Group
	studentsGroup singleflight.Group
)

// 1. 创建学生
func CreateStudent(c *gin.Context) {
	var student models.Student
	if err := c.ShouldBindJSON(&student); err != nil {
		log.Printf("创建学生请求解析失败: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的请求格式"})
		return
	}

	// // 执行数据库操作
	// sqlstr := "INSERT INTO students (id, name, tel, study) VALUES (?, ?, ?, ?)"
	// result, err := database.DB.Exec(sqlstr, student.ID, student.Name, student.Tel, student.Study)
	// if err != nil {
	// 	log.Printf("创建学生失败: %v, 数据: %+v", err, student)
	// 	c.JSON(http.StatusInternalServerError, gin.H{"error": "创建学生失败"})
	// 	return
	// }
	// // 获取影响行数
	// rowsAffected, _ := result.RowsAffected()
	// log.Printf("成功创建学生 %s, 影响行数: %d", student.ID, rowsAffected)

	// GORM 创建信息
	// func (db *gorm.DB) Create(value interface{}) (tx *gorm.DB)
	result := database.GormDB.Create(&student)
	if result.Error != nil {
		log.Printf("创建学生失败: %v, 数据: %+v", result.Error, student)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建学生失败"})
		return
	}
	log.Printf("成功创建学生 %s", student.ID)

	// 清除学生列表缓存！！！
	ctx := context.Background()
	// func (c redis.cmdable) Del(ctx context.Context, keys ...string) *redis.IntCmd
	if err := database.RedisClient.Del(ctx, "students:list").Err(); err != nil {
		log.Printf("清除列表缓存失败: %v", err)
	}

	c.JSON(http.StatusCreated, student)
}

// 2. 获取所有学生
func ListStudents(c *gin.Context) {
	ctx := context.Background()
	cacheKey := "students:list"

	// 从Redis获取列表缓存
	// func (c redis.cmdable) Get(ctx context.Context, key string) *redis.StringCmd
	// cachedData, err := database.RedisClient.Get(ctx, cacheKey).Result()
	// if err == nil {
	// 	var students []models.Student
	// 	if err := json.Unmarshal([]byte(cachedData), &students); err == nil {
	// 		log.Printf("从缓存获取学生列表 (数量: %d)", len(students))
	// 		c.JSON(http.StatusOK, students)
	// 		return
	// 	}
	// 	log.Printf("列表缓存数据解析失败: %v", err)
	// }

	// // 没获取到数据，继续在数据库中查询
	// sqlstr := "SELECT id, name, tel, study FROM students"
	// rows, err := database.DB.Query(sqlstr)
	// if err != nil {
	// 	log.Printf("查询学生列表失败: %v", err)
	// 	c.JSON(http.StatusInternalServerError, gin.H{"error": "获取学生列表失败"})
	// 	return
	// }
	// defer rows.Close()

	// var students []models.Student
	// for rows.Next() {
	// 	var temp models.Student
	// 	if err := rows.Scan(&temp.ID, &temp.Name, &temp.Tel, &temp.Study); err != nil {
	// 		log.Printf("解析学生数据失败: %v", err)
	// 		continue
	// 	}
	// 	students = append(students, temp)
	// }
	// // 检查迭代错误是一个良好的习惯   嘻嘻
	// if err := rows.Err(); err != nil {
	// 	log.Printf("遍历学生数据出错: %v", err)
	// }

	// // GORM 查询所有记录
	// var students []models.Student
	// // func (db *gorm.DB) Find(dest interface{}, conds ...interface{}) (tx *gorm.DB)
	// result := database.GormDB.Find(&students)
	// if result.Error != nil {
	// 	log.Printf("查询学生列表失败: %v", result.Error)
	// 	c.JSON(http.StatusInternalServerError, gin.H{"error": "获取学生列表失败"})
	// 	return
	// }

	// log.Printf("从数据库获取学生列表 (数量: %d)", len(students))

	// // 存入缓存 (设置2分钟过期)
	// studentsJSON, _ := json.Marshal(students)
	// // func (c redis.cmdable) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) *redis.StatusCmd
	// if err := database.RedisClient.Set(ctx, cacheKey, studentsJSON, 2*time.Minute).Err(); err != nil {
	// 	log.Printf("缓存学生列表失败: %v", err)
	// }
	// c.JSON(http.StatusOK, students)

	studentListResult, err, _ := studentsGroup.Do(cacheKey, func() (interface{}, error) {
		cacheData, err := database.RedisClient.Get(ctx, cacheKey).Result()
		if err == nil {
			if cacheData == "" {
				log.Printf("空缓存: %s", cacheKey)
				return []models.Student{}, nil
			}
			var studentTemp []models.Student
			if err := json.Unmarshal([]byte(cacheData), &studentTemp); err == nil {
				log.Printf("从缓存获取学生列表 (数量: %d)", len(studentTemp))
				return studentTemp, nil
			}
			log.Printf("列表缓存解析失败: %v", err)
		} else if err != redis.Nil {
			log.Printf("Redis 获取失败: %v", err)
		}

		var studentTemp []models.Student
		studentResult := database.GormDB.Find(&studentTemp)
		if studentResult.Error != nil {
			log.Printf("查询学生列表失败: %v", studentResult.Error)
			return nil, studentResult.Error
		}
		log.Printf("从数据库获取学生列表 (数量: %d)", len(studentTemp))

		expiration := 10 * time.Minute
		if len(studentTemp) == 0 {
			expiration = 5 * time.Second
		}
		allStudentJson, err := json.Marshal(studentTemp)
		if err != nil {
			log.Printf("序列化学生列表失败: %v", err)
			return studentTemp, nil
		}
		if err := database.RedisClient.Set(ctx, cacheKey, allStudentJson, expiration).Err(); err != nil {
			log.Printf("Redis 写入失败: %s, 错误: %v", cacheKey, err)
		} else {
			log.Printf("Redis 写入成功: %s", cacheKey)
		}
		return studentTemp, nil
	})

	if err != nil {
		log.Printf("获取全部学生信息失败,err = %s", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, studentListResult)
}

// 3. 获取单个学生
func GetStudent(c *gin.Context) {
	id := c.Param("id")
	ctx := context.Background()
	cacheKey := "student:" + id
	studentInfomation, err, _ := studentGroup.Do(cacheKey, func() (interface{}, error) {
		cachedData, err := database.RedisClient.Get(ctx, cacheKey).Result()
		if err == nil {
			if cachedData == "" {
				log.Printf("空缓存: %s", cacheKey)
				return nil, errors.New("空缓存")
			}
			var studentTemp models.Student
			var e error
			if e = json.Unmarshal([]byte(cachedData), &studentTemp); e == nil {
				log.Printf("从缓存获取学生 %s", id)
				return studentTemp, nil
			}
			log.Printf("缓存解析失败: %v", e)
		} else if err != redis.Nil {
			log.Printf("Redis 获取失败: %v", err)
		}

		var studentTemp models.Student
		result := database.GormDB.First(&studentTemp, "id = ?", id)
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			log.Printf("学生不存在: %s", id)
			return nil, errors.New("数据库中学生信息不存在")
		}
		if result.Error != nil {
			log.Printf("数据库查询失败: %v", result.Error)
			return nil, result.Error
		}
		log.Printf("从数据库获取学生 %s", id)

		studentJSON, err := json.Marshal(studentTemp)
		if err != nil {
			log.Printf("序列化学生失败: %v", err)
			return studentTemp, nil
		}

		if err := database.RedisClient.Set(ctx, cacheKey, studentJSON, 5*time.Minute).Err(); err != nil {
			log.Printf("Redis 写入失败: %s, 错误: %v", cacheKey, err)
		} else {
			log.Printf("Redis 写入成功: %s", cacheKey)
		}
		return studentTemp, nil
	})

	if err != nil {
		if err.Error() == "数据库中学生信息不存在" {
			c.JSON(http.StatusNotFound, gin.H{"error": "学生信息不存在"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取学生信息错误"})
		return
	}
	c.JSON(http.StatusOK, studentInfomation)
}

// 4. 更新学生 (更新后清除缓存)
func UpdateStudent(c *gin.Context) {
	id := c.Param("id")
	var updateData struct {
		Name  string `json:"name"`
		Tel   string `json:"tel"`
		Study string `json:"study"`
	}
	if err := c.ShouldBindJSON(&updateData); err != nil {
		log.Printf("更新请求解析失败: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的请求格式"})
		return
	}

	// sqlstr := "UPDATE students SET name = ?, tel = ?, study = ? WHERE id = ?"
	// result, err := database.DB.Exec(sqlstr, updateData.Name, updateData.Tel, updateData.Study, id)
	// if err != nil {
	// 	log.Printf("更新学生 %s 失败: %v", id, err)
	// 	c.JSON(http.StatusInternalServerError, gin.H{"error": "更新学生失败"})
	// 	return
	// }

	// rowsAffected, err := result.RowsAffected()
	// if err != nil {
	// 	log.Printf("获取影响行数失败: %v", err)
	// }
	// if rowsAffected == 0 {
	// 	log.Printf("更新失败: 学生不存在 %s", id)
	// 	c.JSON(http.StatusNotFound, gin.H{"error": "学生不存在"})
	// 	return
	// }
	// log.Printf("成功更新学生 %s, 影响行数: %d", id, rowsAffected)

	// GORM 更新记录
	// 先查再更新
	var student models.Student
	result := database.GormDB.First(&student, "id = ?", id)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		log.Printf("更新失败: 学生不存在 %s", id)
		c.JSON(http.StatusNotFound, gin.H{"error": "学生不存在"})
		return
	}
	if result.Error != nil {
		log.Printf("查询学生 %s 失败: %v", id, result.Error)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新学生失败"})
		return
	}

	// 更新字段 (仅更新非空字段)
	updates := map[string]interface{}{}
	if updateData.Name != "" {
		updates["name"] = updateData.Name
	}
	if updateData.Tel != "" {
		updates["tel"] = updateData.Tel
	}
	if updateData.Study != "" {
		updates["study"] = updateData.Study
	}

	if len(updates) > 0 {
		result = database.GormDB.Model(&student).Updates(updates) // 更新指定字段
		if result.Error != nil {
			log.Printf("更新学生 %s 失败: %v", id, result.Error)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "更新学生失败"})
			return
		}
		log.Printf("成功更新学生 %s", id)
	} else {
		log.Printf("没有需要更新的字段 for 学生 %s", id)
	}

	// 清除该学生的缓存
	ctx := context.Background()
	cacheKey := "student:" + id
	if err := database.RedisClient.Del(ctx, cacheKey).Err(); err != nil {
		log.Printf("清除缓存失败: %v", err)
	} else {
		log.Printf("已清除学生缓存: %s", cacheKey)
	}
	if err := database.RedisClient.Del(ctx, "students:list").Err(); err != nil {
		log.Printf("清除列表缓存失败: %v", err)
	}

	c.JSON(http.StatusOK, gin.H{"status": "更新成功"})
}

// 5. 删除学生 (清除缓存)
func DeleteStudent(c *gin.Context) {
	id := c.Param("id")
	// sqlstr := "DELETE FROM students WHERE id = ?"
	// result, err := database.DB.Exec(sqlstr, id)
	// if err != nil {
	// 	log.Printf("删除学生 %s 失败: %v", id, err)
	// 	c.JSON(http.StatusInternalServerError, gin.H{"error": "删除学生失败"})
	// 	return
	// }

	// rowsAffected, err := result.RowsAffected()
	// if err != nil {
	// 	log.Printf("获取影响行数失败: %v", err)
	// }

	// if rowsAffected == 0 {
	// 	log.Printf("删除失败: 学生不存在 %s", id)
	// 	c.JSON(http.StatusNotFound, gin.H{"error": "学生不存在"})
	// 	return
	// }
	// log.Printf("成功删除学生 %s, 影响行数: %d", id, rowsAffected)

	// GORM 删除记录
	// 先查再删
	var student models.Student
	result := database.GormDB.First(&student, "id = ?", id)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		log.Printf("删除失败: 学生不存在 %s", id)
		c.JSON(http.StatusNotFound, gin.H{"error": "学生不存在"})
		return
	}
	if result.Error != nil {
		log.Printf("查询学生 %s 失败: %v", id, result.Error)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除学生失败"})
		return
	}

	result = database.GormDB.Delete(&student)
	if result.Error != nil {
		log.Printf("删除学生 %s 失败: %v", id, result.Error)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除学生失败"})
		return
	}
	log.Printf("成功删除学生 %s", id)

	// 清除该学生的缓存
	ctx := context.Background()
	cacheKey := "student:" + id
	if err := database.RedisClient.Del(ctx, cacheKey).Err(); err != nil {
		log.Printf("清除缓存失败: %v", err)
	} else {
		log.Printf("已清除学生缓存: %s", cacheKey)
	}
	if err := database.RedisClient.Del(ctx, "students:list").Err(); err != nil {
		log.Printf("清除列表缓存失败: %v", err)
	}

	c.JSON(http.StatusOK, gin.H{"status": "删除成功"})
}
