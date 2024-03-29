package handlers

import (
	"context"
	"log"
	"lost-item/cloud/googlecloud"
	"lost-item/database"
	"lost-item/database/postgresd"
	"lost-item/model"
	"net/http"
	"strconv"

	"github.com/google/uuid"
	imgupload "github.com/olahol/go-imageupload"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	db    database.DBConn
	cloud googlecloud.GCloud
}

func (h *Handler) Init() {
	var err error
	if h.db, err = postgresd.NewPostgresd(); err != nil {
		log.Fatalf("Database connection failed")
	}
	h.db.CreateTable()

	ctx := context.Background()
	x, err := googlecloud.NewGoogleCloud(ctx)
	if err != nil {
		log.Fatalf("Cloud initialization failure")
	}
	h.cloud = *x
}

func (h Handler) Search(c *gin.Context) {
	var request struct {
		Location1 model.Location `json:"location1" binding:"required"`
		Location2 model.Location `json:"location2" binding:"required"`

		Query string   `json:"query"`
		Tags  []string `json:"tags"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		log.Println("うまくバインドできませんでした")
		c.Status(http.StatusBadRequest)
		return
	}
	search_result, err := h.db.Search(request.Location1, request.Location2, request.Query, request.Tags)

	if err != nil {
		c.Status(http.StatusBadRequest)
		log.Println(err)
		return
	}

	for _, v := range search_result.Items {
		if "" == v.Kinds[0] {
			search_result.Items[0].Kinds = make([]string, 0)
		}
	}

	c.JSON(http.StatusOK, search_result)
}

func (h Handler) ItemDetail(c *gin.Context) {
	item_id, err := strconv.ParseInt(c.Param("id"), 10, 64)

	if err != nil {
		c.String(http.StatusBadRequest, "Bad Request")
		return
	}

	item_detail, err := h.db.ItemDetail(uint64(item_id))
	if err != nil {
		c.Status(http.StatusBadRequest)
		log.Println(err)
		return
	}

	if "" == item_detail.Kinds[0] {
		item_detail.Kinds = make([]string, 0)
	}

	c.JSON(http.StatusOK, item_detail)
}

func (h Handler) RegisterItem(c *gin.Context) {
	var register_item model.LostItem
	var err error

	err = c.BindJSON(&register_item)
	if err != nil {
		log.Println("バインドに失敗しました")
		c.Status(http.StatusBadRequest)
		return
	}

	register_item, err = h.db.InsertItem(register_item)
	if err != nil {
		c.Status(http.StatusBadRequest)
		return
	}

	c.JSON(http.StatusOK, register_item)

}

func (h Handler) DeleteItem(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.Status(http.StatusBadRequest)
		return
	}
	h.db.CompleteItem(id)

	c.Status(http.StatusOK)
}

func (h Handler) Parse(c *gin.Context) {
	img, err := imgupload.Process(c.Request, "file")
	if err != nil {
		c.Status(http.StatusBadRequest)
		return
	}

	uuid, err := uuid.NewRandom()
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}
	filename := uuid.String()

	if img.ContentType == "image/png" {
		filename += ".png"
	} else if img.ContentType == "image/jpeg" {
		filename += ".jpg"
	} else {
		c.Status(http.StatusBadRequest)
		return
	}

	if err := h.cloud.UploadImage(img.Data, filename); err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}

	var img_info model.ImageInfo
	if img_info.ImageURL, err = h.cloud.GetURL(filename); err != nil {
		log.Println("URLを取得できませんでした")
		c.Status(http.StatusInternalServerError)
		return
	}

	objects, err := h.cloud.ObjectRecognition(filename)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}

	labels, err := h.cloud.LabelRecognition(filename)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}

	objects = append(objects, labels...)

	// mapには重複したキーをセットできないことを利用して重複を取り除いている
	m := make(map[string]bool)
	for _, v := range objects {
		if !m[v] {
			m[v] = true
			img_info.Kinds = append(img_info.Kinds, v)
		}
	}

	c.JSON(http.StatusOK, img_info)
}

func (h Handler) UpdateItem(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.Status(http.StatusBadRequest)
		return
	}
	var update_item model.UpdateLostItem

	err = c.BindJSON(&update_item)
	if err != nil {
		log.Println("バインドに失敗しました")
		c.Status(http.StatusBadRequest)
		return
	}

	updated_item, err := h.db.UpdateItem(id, update_item)
	if err != nil {
		c.Status(http.StatusBadRequest)
		return
	}

	c.JSON(http.StatusOK, updated_item)
}
