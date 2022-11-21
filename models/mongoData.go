package models

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
	"smartest-go/pkg/app"
	"smartest-go/pkg/e"
	"smartest-go/pkg/mongo"
	utilconst "smartest-go/pkg/util/const"
)

type MongoList struct {
	ObjectId    string                 `form:"_id,omitempty" json:"_id,omitempty"`
	FDataBase   string                 `form:"f_database,omitempty"  json:"f_database,omitempty"`
	FCollection string                 `form:"f_collection,omitempty" json:"f_collection,omitempty"`
	PageNum     int64                  `form:"pagenum,omitempty"    json:"pagenum,omitempty"`
	PageSize    int64                  `form:"pagesize,omitempty"   json:"pagesize,omitempty"`
	Filter      map[string]interface{} `form:"filter,omitempty" json:"filter,omitempty"`
	Aggregate   interface{}            `form:"aggregate,omitempty" json:"aggregate,omitempty"`
	Update      map[string]interface{} `form:"update,omitempty" json:"update,omitempty"`
	Column      map[string]interface{} `form:"column,omitempty" json:"column,omitempty"`
	Sort        map[string]interface{} `form:"sort,omitempty"   json:"sort,omitempty"`
}

// MongoUpdateFunc mongoDataBase update
func MongoUpdateFunc(context *gin.Context) {
	//BindAndValid
	copyReq := &MongoList{}
	err := app.BindAndValid(context, copyReq)
	if err != nil {
		app.ErrorResp(context, e.InvalidParams, err.Error(), nil)
		context.Abort()
		return
	}
	context.Set(utilconst.REQUEST_KEY, copyReq)

	//VerifyReq
	req := context.MustGet(utilconst.REQUEST_KEY).(*MongoList)

	//Check Mongo Collection
	if req.FCollection == "" {
		app.ErrorResp(context, e.ERROR, "UpdateSkillReport err: invalid param collection!", nil)
		context.Abort()
		return
	} else {
		//Check Mongo DataBase
		if req.FDataBase != "" {
			ReporterDB.DB = req.FDataBase
		}

		//Format ObjectId
		if ObjectId, ok := req.Filter["_id"]; ok {
			objectId, err := primitive.ObjectIDFromHex(mongo.GetInterfaceToString(ObjectId))
			if err != nil {
				app.ErrorResp(context, e.ERROR, fmt.Sprintf("ReporterDB.MongoUpdateMany.primitive.ObjectIDFromHex err:%s", err), nil)
				context.Abort()
				return
			}
			req.Filter = bson.M{"_id": objectId}
		}

		data, err := ReporterDB.MongoUpdateMany(req.FCollection, req.Filter, req.Update)

		if err != nil {
			app.ErrorResp(context, e.ERROR, fmt.Sprintf("ReporterDB.MongoUpdateMany err:%s", err), nil)
			context.Abort()
			return
		}
		app.SuccessResp(context, struct {
			Count int64       `json:"count"`
			Data  interface{} `json:"data"`
		}{
			Count: data.ModifiedCount,
			Data:  data,
		})
	}
}

// MongoListFuncFind mongoDataBase find
func MongoListFuncFind(context *gin.Context) {
	//BindAndValid
	copyReq := &MongoList{}
	err := app.BindAndValid(context, copyReq)
	if err != nil {
		app.ErrorResp(context, e.InvalidParams, err.Error(), nil)
		context.Abort()
		return
	}
	context.Set(utilconst.REQUEST_KEY, copyReq)

	//VerifyReq
	req := context.MustGet(utilconst.REQUEST_KEY).(*MongoList)

	//Check Mongo Collection
	if req.FCollection == "" {
		app.ErrorResp(context, e.ERROR, "ListSkillReport err: invalid param collection!", nil)
		context.Abort()
		return
	} else {
		//Check Mongo DataBase
		if req.FDataBase != "" {
			ReporterDB.DB = req.FDataBase
		}

		//Delete ObjectId
		if _, ok := req.Filter["_id"]; ok {
			delete(req.Filter, "_id")
		}

		//Format ObjectId
		if req.ObjectId != "" {
			objectId, err := primitive.ObjectIDFromHex(mongo.GetInterfaceToString(req.ObjectId))
			if err != nil {
				app.ErrorResp(context, e.ERROR, fmt.Sprintf("ReporterDB.MongoFind.primitive.ObjectIDFromHex err:%s", err), nil)
				context.Abort()
				return
			}
			req.Filter = bson.M{"_id": objectId}
		}

		//Format FindOptions
		option := options.Find()
		if req.PageNum < 1 || req.PageSize < 1 {
			req.PageNum, req.PageSize = 1, 10
		}
		option.SetSkip((req.PageNum - 1) * req.PageSize)
		option.SetLimit(req.PageSize)
		if req.Column != nil {
			option.SetProjection(req.Column)
		}
		if req.Sort != nil {
			option.SetSort(req.Sort)
		}

		//MongoFind
		data, err := ReporterDB.MongoFind(req.FCollection, req.Filter, option)
		if err != nil {
			app.ErrorResp(context, e.ERROR, fmt.Sprintf("ReporterDB.MongoFind err:%s", err), nil)
			context.Abort()
			return
		}

		//CountResult
		total, err := ReporterDB.MongoCount(req.FCollection, req.Filter)
		if err != nil {
			app.ErrorResp(context, e.ERROR, fmt.Sprintf("ReporterDB.MongoCount err:%s", err), nil)
			context.Abort()
			return
		}
		// bson to list[map]
		var responseData []map[string]interface{}
		for _, d := range data {
			responseData = append(responseData, d.Map())
		}

		app.SuccessResp(context, struct {
			Count int64                    `json:"count"`
			Data  []map[string]interface{} `json:"data"`
		}{
			Count: total,
			Data:  responseData,
		})
	}
}

// MongoListFuncAggregate mongoDataBase Aggregate
func MongoListFuncAggregate(context *gin.Context) {
	//BindAndValid
	copyReq := &MongoList{}
	err := app.BindAndValid(context, copyReq)
	if err != nil {
		app.ErrorResp(context, e.InvalidParams, err.Error(), nil)
		context.Abort()
		return
	}
	context.Set(utilconst.REQUEST_KEY, copyReq)

	//VerifyReq
	req := context.MustGet(utilconst.REQUEST_KEY).(*MongoList)

	//Check Mongo Collection
	if req.FCollection == "" {
		app.ErrorResp(context, e.ERROR, "ListSkillReport err: invalid param collection!", nil)
		context.Abort()
		return
	} else {
		//Check Mongo DataBase
		if req.FDataBase != "" {
			ReporterDB.DB = req.FDataBase
		}

		//Format FindOptions
		option := options.Aggregate()

		//MongoAggregate
		data, err := ReporterDB.MongoAggregate(req.FCollection, req.Aggregate, option)
		if err != nil {
			app.ErrorResp(context, e.ERROR, fmt.Sprintf("ReporterDB.MongoFind err:%s", err), nil)
			context.Abort()
			return
		}

		//CountResult
		// bson to list[map]
		var responseData []map[string]interface{}
		for _, d := range data {
			responseData = append(responseData, d.Map())
		}

		app.SuccessResp(context, struct {
			Count int64                    `json:"count"`
			Data  []map[string]interface{} `json:"data"`
		}{
			Count: 0,
			Data:  responseData,
		})
	}
}

func MongoListAndExportFunc(context *gin.Context) {
	//BindAndValid
	copyReq := &MongoList{}
	err := app.BindAndValid(context, copyReq)
	if err != nil {
		app.ErrorResp(context, e.InvalidParams, err.Error(), nil)
		context.Abort()
		return
	}
	context.Set(utilconst.REQUEST_KEY, copyReq)

	//VerifyReq
	req := context.MustGet(utilconst.REQUEST_KEY).(*MongoList)

	//Check Mongo Collection
	if req.FCollection == "" {
		app.ErrorResp(context, e.ERROR, "ListSkillReport err: invalid param collection!", nil)
		context.Abort()
		return
	} else {
		//Check Mongo DataBase
		if req.FDataBase != "" {
			ReporterDB.DB = req.FDataBase
		}

		//Delete ObjectId
		if _, ok := req.Filter["_id"]; ok {
			delete(req.Filter, "_id")
		}

		//Format ObjectId
		if req.ObjectId != "" {
			objectId, err := primitive.ObjectIDFromHex(mongo.GetInterfaceToString(req.ObjectId))
			if err != nil {
				app.ErrorResp(context, e.ERROR, fmt.Sprintf("ReporterDB.MongoFind.primitive.ObjectIDFromHex err:%s", err), nil)
				context.Abort()
				return
			}
			req.Filter = bson.M{"_id": objectId}
		}

		//Format FindOptions
		option := options.Find()
		if req.PageNum < 1 || req.PageSize < 1 {
			req.PageNum, req.PageSize = 1, 10
		}
		option.SetSkip((req.PageNum - 1) * req.PageSize)
		option.SetLimit(req.PageSize)
		if req.Column != nil {
			option.SetProjection(req.Column)
		}
		if req.Sort != nil {
			option.SetSort(req.Sort)
		}

		//MongoFind
		data, err := ReporterDB.MongoFind(req.FCollection, req.Filter, option)
		if err != nil {
			app.ErrorResp(context, e.ERROR, fmt.Sprintf("ReporterDB.MongoFind err:%s", err), nil)
			context.Abort()
			return
		}

		//CountResult
		total, err := ReporterDB.MongoCount(req.FCollection, req.Filter)
		if err != nil {
			app.ErrorResp(context, e.ERROR, fmt.Sprintf("ReporterDB.MongoCount err:%s", err), nil)
			context.Abort()
			return
		}
		// bson to list[map]
		var responseData []map[string]interface{}
		for _, d := range data {
			responseData = append(responseData, d.Map())
		}

		filename := ExportExcelByMap("export", responseData)
		app.SuccessResp(context, struct {
			Count int64  `json:"count"`
			Data  string `json:"data"`
		}{
			Count: total,
			Data:  filename,
		})
	}
}
