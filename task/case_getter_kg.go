package task

import (
	"context"
	"fmt"
	"github.com/360EntSecGroup-Skylar/excelize"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
	"math/rand"
	"smartest-go/pkg/mongo"
	"strconv"
	"strings"
	"sync"
	"time"
)

/*
用于知识图谱的用例收集
*/

var (
	entityRLTable   = "entity_rl"
	entityTable     = "entity"
	ontologyRLTable = "ontology_rl"
)

func (KG *KGTask) CaseGetterKG(c context.Context) {
	KG.KGCaseGetterMongo = KG.KGDataSourceConfig.KGDataBase
	KG.KGCaseGetterMongo.MongoPoolConnect(10000)
	switch KG.KGDataSourceConfig.CType {
	case 1:
		if KG.KGDataSourceConfig.IsRandom == "yes" {
			KG.fakeQuerySingleStepRandomly(c)
		} else {
			KG.mockQueryOneStep(c)
		}
	case 2:
		KG.mockQueryTwoStep(c)
	}
}

func (KG *KGTask) fakeQuerySingleStepRandomly(ctx context.Context) {
	// 抽取关系 从关系表中 随机抽取n条关系
	caseListRl, _ := KG.KGCaseGetterMongo.MongoAggregate(entityRLTable, []bson.M{
		{"$match": bson.M{"status": bson.M{"$lt": 2}, "is_del": false}},
		{"$sample": bson.M{"size": KG.KGDataSourceConfig.CaseNum / 5}}})
	var wg sync.WaitGroup
	wg.Add(1)
	go func(ctx context.Context) {
		defer wg.Done()
		for _, i := range caseListRl {
			select {
			case <-ctx.Done():
				return
			default:
				nowCaseNum := len(KG.req)
				totalCaseNum := int(KG.KGDataSourceConfig.CaseNum)
				if nowCaseNum == totalCaseNum {
					return
				}
				if r := KG.fakeQuerySingleStep(i); r != nil {
					KG.req = append(KG.req, r)
					if value, ok := taskInfoMap[KG.KGConfig.TaskName]; ok {
						value.ProgressPercent = nowCaseNum * 100 / totalCaseNum
						value.Progress = fmt.Sprintf(`%d/%d`, nowCaseNum, totalCaseNum)
					}
				}
			}
		}
	}(ctx)
	wg.Wait()
	select {
	case <-ctx.Done():
		return
	default:
		if len(KG.req) < int(KG.KGDataSourceConfig.CaseNum) {
			KG.fakeQuerySingleStepRandomly(ctx)
		} else {
			KG.req = KG.req[:KG.KGDataSourceConfig.CaseNum]
		}
	}
}

func (KG *KGTask) fakeQuerySingleStep(entityRl *bson.D) (Req *KGTaskReq) {
	// 单跳用例构造  <entityA>的<relation1>是<entityB>
	var infoEID1, infoEID2, infoOTRLID []*bson.D
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		infoEID1, _ = KG.KGCaseGetterMongo.MongoFind(entityTable, bson.M{"_id": entityRl.Map()["e_id"], "need_audit": false}, options.Find().SetLimit(1))
		wg.Done()
	}()
	wg.Add(1)
	go func() {
		infoEID2, _ = KG.KGCaseGetterMongo.MongoFind(entityTable, bson.M{"_id": entityRl.Map()["e_id2"], "need_audit": false}, options.Find().SetLimit(1))
		wg.Done()
	}()
	wg.Add(1)
	go func() {
		infoOTRLID, _ = KG.KGCaseGetterMongo.MongoFind(ontologyRLTable, bson.M{"_id": entityRl.Map()["ot_rl_id"]}, options.Find().SetLimit(1))
		wg.Done()
	}()
	wg.Wait()

	if infoEID1 != nil && infoOTRLID != nil && infoEID2 != nil {
		Req = &KGTaskReq{
			Query:        mongo.GetInterfaceToString(infoEID1[0].Map()["name"]) + "的" + mongo.GetInterfaceToString(infoOTRLID[0].Map()["name"]),
			ExpectAnswer: mongo.GetInterfaceToString(infoEID2[0].Map()["name"]),
		}
	}
	return
}

// KGTemplate 模板JSON文件结构
type KGTemplate struct {
	Relation     string        `json:"relation,omitempty"`
	Query        string        `json:"query,omitempty"`
	ExpectAnswer string        `json:"expect_answer,omitempty"`
	Model        []*KGTemplate `json:"model,omitempty"`
}

func replaceSlot(text, entityA, entityB, entityC string) string {
	if entityA != "" {
		text = strings.ReplaceAll(text, "{A}", entityA)
	}
	if entityB != "" {
		text = strings.ReplaceAll(text, "{B}", entityB)
	}
	if entityC != "" {
		text = strings.ReplaceAll(text, "{C}", entityC)
	}
	return text
}

func (KG *KGTask) returnOneTemplate(tmpList []*KGTemplate) *KGTemplate {
	var mu sync.Mutex
	mu.Lock()
	rand.Seed(time.Now().UnixNano())
	tmp := tmpList[rand.Intn(len(tmpList))]
	mu.Unlock()
	return tmp
}

func (KG *KGTask) mockQueryTwoStepByTemplate() (Req []*KGTaskReq) {
	// 两跳用例构造  <A> <relation1> <B> <relation2> <C>
	// 两跳用例构造  周杰伦   母亲    叶惠美    配偶    周耀中

	// 随机抽一条模板出来
	tmp1 := KG.returnOneTemplate(KG.KGDataSourceConfig.TemplateJson)
	Relation1 := tmp1.Relation
	tmp2 := KG.returnOneTemplate(tmp1.Model)
	Relation2 := tmp2.Relation

	// 根据关系1的中文名查本体与本体之间的关系列表ids
	Relation1IDs, _ := KG.KGCaseGetterMongo.MongoAggregate(ontologyRLTable, []bson.M{
		{"$match": bson.M{"name": Relation1}},
		{"$sample": bson.M{"size": KG.KGDataSourceConfig.CaseNum / 5}}})
	if Relation1IDs == nil {
		return
	}
	// 遍历关系1的ids
	for _, Relation1ID := range Relation1IDs {

		// 根据关系1的id查询三元组triplets(e_id, e_id2, ot_rl_id)
		triplets, _ := KG.KGCaseGetterMongo.MongoFind(entityRLTable, bson.M{"ot_rl_id": Relation1ID.Map()["_id"], "status": bson.M{"$lt": 2}, "is_del": false})
		for _, triplet := range triplets {
			var wg sync.WaitGroup
			var Triplets2, EntityA, EntityB, EntityC []*bson.D // 三个实体对应的mongo数据信息
			var A, B, C string                                 // 三个实体的中文名
			var EntityAid, EntityBid, EntityCid interface{}    // 三个实体的id

			EntityAid = triplet.Map()["e_id"]
			EntityBid = triplet.Map()["e_id2"]
			wg.Add(1)
			go func() {
				Triplets2, _ = KG.KGCaseGetterMongo.MongoFind(entityRLTable, bson.M{"e_id": EntityBid, "status": bson.M{"$lt": 2}, "is_del": false})
				wg.Done()
			}()
			wg.Add(1)
			go func() {
				EntityA, _ = KG.KGCaseGetterMongo.MongoFind(entityTable, bson.M{"_id": EntityAid, "need_audit": false}, options.Find().SetLimit(1))
				wg.Done()
			}()
			wg.Add(1)
			go func() {
				EntityB, _ = KG.KGCaseGetterMongo.MongoFind(entityTable, bson.M{"_id": EntityBid, "need_audit": false}, options.Find().SetLimit(1))
				wg.Done()
			}()
			wg.Wait()

			// 异常捕获
			if EntityA == nil || EntityB == nil {
				continue
			}
			A = mongo.GetInterfaceToString(EntityA[0].Map()["name"])
			B = mongo.GetInterfaceToString(EntityB[0].Map()["name"])

			// 查到mongo数据后 看是否需要去nebula数据校对
			if result1 := KG.verifyNebula(EntityAid, Relation1, B, KG.KGDataSourceConfig.VerifyAddr); !result1 {
				continue
			}

			//	根据eid2 找第二个三元组关系
			for _, x := range Triplets2 {
				EntityCid = x.Map()["e_id2"]
				var kk []*bson.D
				wg.Add(1)
				go func() {
					kk, _ = KG.KGCaseGetterMongo.MongoFind(ontologyRLTable, bson.M{"_id": x.Map()["ot_rl_id"]})
					wg.Done()
				}()
				wg.Add(1)
				go func() {
					EntityC, _ = KG.KGCaseGetterMongo.MongoFind(entityTable, bson.M{"_id": EntityCid, "need_audit": false}, options.Find().SetLimit(1))
					wg.Done()
				}()
				wg.Wait()

				if EntityC == nil {
					continue
				}
				C = mongo.GetInterfaceToString(EntityC[0].Map()["name"])

				if result2 := KG.verifyNebula(EntityBid, Relation2, C, KG.KGDataSourceConfig.VerifyAddr); !result2 {
					continue
				}
				if kk == nil {
					continue
				}
				if kk[0].Map()["name"] != Relation2 {
					continue
				}
				for _, tmp3 := range tmp2.Model {
					Req = append(Req, &KGTaskReq{
						Query:        replaceSlot(tmp3.Query, A, B, C),
						ExpectAnswer: replaceSlot(tmp3.ExpectAnswer, A, B, C),
					})
				}
			}
		}
	}
	return
}

func (KG *KGTask) verifyNebula(entityId interface{}, relation, expectAnswer, env string) bool {
	// 判断用例数据是否需要在nebula中校验有效性
	if KG.KGDataSourceConfig.IsVerify == "yes" && KG.KGDataSourceConfig.VerifyAddr != "" {
		// TODO
		return true
	}
	return true
}

func (KG *KGTask) mockQueryOneStepByTemplate() (Req []*KGTaskReq) {
	// 单跳用例构造  <A> <relation1> <B>
	// 单跳用例构造  周杰伦   母亲    叶惠美

	// 随机抽一条模板出来
	tmp1 := KG.returnOneTemplate(KG.KGDataSourceConfig.TemplateJson)
	Relation1 := tmp1.Relation

	// 根据关系1的中文名查本体与本体之间的关系列表ids
	Relation1IDs, _ := KG.KGCaseGetterMongo.MongoAggregate(ontologyRLTable, []bson.M{
		{"$match": bson.M{"name": Relation1}},
		{"$sample": bson.M{"size": KG.KGDataSourceConfig.CaseNum / 5}}})

	// 遍历关系1的ids
	for _, Relation1ID := range Relation1IDs {

		// 根据关系1的id查询三元组triplets(e_id, e_id2, ot_rl_id)
		triplets, _ := KG.KGCaseGetterMongo.MongoFind(entityRLTable, bson.M{"ot_rl_id": Relation1ID.Map()["_id"], "status": bson.M{"$lt": 2}, "is_del": false})
		for _, triplet := range triplets {
			var wg sync.WaitGroup
			var EntityB, EntityA []*bson.D
			wg.Add(1)
			go func() {
				EntityB, _ = KG.KGCaseGetterMongo.MongoFind(entityTable, bson.M{"_id": triplet.Map()["e_id2"], "need_audit": false}, options.Find().SetLimit(1))
				wg.Done()
			}()
			wg.Add(1)
			go func() {
				EntityA, _ = KG.KGCaseGetterMongo.MongoFind(entityTable, bson.M{"_id": triplet.Map()["e_id"], "need_audit": false}, options.Find().SetLimit(1))
				wg.Done()
			}()
			wg.Wait()
			if EntityB != nil && EntityA != nil {
				A := mongo.GetInterfaceToString(EntityA[0].Map()["name"])
				B := mongo.GetInterfaceToString(EntityB[0].Map()["name"])
				for _, tmp2 := range tmp1.Model {
					Req = append(Req, &KGTaskReq{
						Query:        replaceSlot(tmp2.Query, A, B, ""),
						ExpectAnswer: replaceSlot(tmp2.ExpectAnswer, A, B, ""),
					})
				}
			}
		}
	}
	return
}

func (KG *KGTask) mockQueryTwoStep(ctx context.Context) {
	var wg sync.WaitGroup
	wg.Add(1)
	go func(ctx context.Context) {
		defer wg.Done()
		select {
		case <-ctx.Done():
			return
		default:
			nowCaseNum := len(KG.req)
			totalCaseNum := int(KG.KGDataSourceConfig.CaseNum)
			if nowCaseNum == totalCaseNum {
				return
			}
			if r := KG.mockQueryTwoStepByTemplate(); r != nil {
				for _, item := range r {
					KG.req = append(KG.req, item)
				}
			}
			if value, ok := taskInfoMap[KG.KGConfig.TaskName]; ok {
				value.ProgressPercent = nowCaseNum * 100 / totalCaseNum
				value.Progress = fmt.Sprintf(`%d/%d`, nowCaseNum, totalCaseNum)
			}
		}
	}(ctx)
	wg.Wait()
	select {
	case <-ctx.Done():
		return
	default:
		if len(KG.req) < int(KG.KGDataSourceConfig.CaseNum) {
			KG.mockQueryTwoStep(ctx)
		} else {
			KG.req = KG.req[:KG.KGDataSourceConfig.CaseNum]
		}
	}
}

func (KG *KGTask) mockQueryOneStep(ctx context.Context) {
	var wg sync.WaitGroup
	wg.Add(1)
	go func(ctx context.Context) {
		defer wg.Done()
		select {
		case <-ctx.Done():
			return
		default:
			nowCaseNum := len(KG.req)
			totalCaseNum := int(KG.KGDataSourceConfig.CaseNum)
			if nowCaseNum == totalCaseNum {
				return
			}
			if r := KG.mockQueryOneStepByTemplate(); r != nil {
				for _, item := range r {
					KG.req = append(KG.req, item)
				}
			}
			if value, ok := taskInfoMap[KG.KGConfig.TaskName]; ok {
				value.ProgressPercent = nowCaseNum * 100 / totalCaseNum
				value.Progress = fmt.Sprintf(`%d/%d`, nowCaseNum, totalCaseNum)
			}
		}
	}(ctx)
	wg.Wait()
	select {
	case <-ctx.Done():
		return
	default:
		if len(KG.req) < int(KG.KGDataSourceConfig.CaseNum) {
			KG.mockQueryOneStep(ctx)
		} else {
			KG.req = KG.req[:KG.KGDataSourceConfig.CaseNum]
		}
	}
}

func ExcelKGReader(fileName, sheetName string) (req []*KGTaskReq) {
	if !strings.Contains(fileName, "./upload/") {
		fileName = "./upload/" + fileName
	}
	f, err := excelize.OpenFile(fileName)
	if err != nil {
		return
	}
	rows := f.GetRows(sheetName)

	tableHeader := make(map[int]string)
	for index, row := range rows {
		if index == 0 {
			// 记录表头
			for i, cellValue := range row {
				tableHeader[i] = cellValue
			}
			continue
		}
		tmpReq := &KGTaskReq{
			Id:           0,
			Query:        "",
			ExpectAnswer: "",
		}
		for i, cellValue := range row {
			// 记录表数据
			if tableHeader[i] == "id" {
				num, _ := strconv.Atoi(cellValue)
				tmpReq.Id = int64(num)
			}
			if tableHeader[i] == "query" {
				tmpReq.Query = cellValue
			}
			if tableHeader[i] == "expect_answer" {
				tmpReq.ExpectAnswer = cellValue
			}
		}
		req = append(req, tmpReq)
	}
	return
}
