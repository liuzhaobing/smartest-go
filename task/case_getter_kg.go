package task

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
	"math/rand"
	"smartest-go/pkg/mongo"
	"sync"
	"time"
)

/*
用于知识图谱的用例收集
*/

var (
	entityRLFilter  = bson.M{"status": 0}
	entityRLTable   = "entity_rl"
	entityTable     = "entity"
	ontologyRLTable = "ontology_rl"
)

func (KG *KGTask) CaseGetterKG() {
	KG.KGCaseGetterMongo = KG.KGDataSourceConfig.KGDataBase
	KG.KGCaseGetterMongo.MongoPoolConnect(10000)
	switch KG.KGDataSourceConfig.CType {
	case 1:
		if KG.KGDataSourceConfig.IsRandom == "yes" {
			KG.fakeQuerySingleStepRandomly()
		}
	case 2:
		if KG.KGDataSourceConfig.IsRandom == "yes" {
			KG.fakeQueryTwoStepRandomly()
		}
	}
}

func (KG *KGTask) fakeQuerySingleStepRandomly() {
	// 抽取关系 从关系表中 随机抽取n条关系
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	caseListRl, _ := KG.KGCaseGetterMongo.MongoAggregate(entityRLTable, []bson.M{
		{"$sample": bson.M{"size": KG.KGDataSourceConfig.CaseNum}},
		{"$match": entityRLFilter}})
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
			KG.fakeQuerySingleStepRandomly()
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

func (KG *KGTask) fakeQueryTwoStepRandomly() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

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
			if r := KG.fakeQueryTwoStep(); r != nil {
				KG.req = append(KG.req, r)
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
			KG.fakeQueryTwoStepRandomly()
		}
	}
}

func (KG *KGTask) fakeQueryTwoStep() (Req *KGTaskReq) {
	// 两跳用例构造  <entityA> <relation1> <entityB> <relation2> <entityC>
	nu := 20

	rl1OntologyName, rl2OntologyName, model := KG.getOneTemplate(KG.KGDataSourceConfig.TemplateJson)

	// 在ontology_rl中找作者属性id
	authorRl, _ := KG.KGCaseGetterMongo.MongoFind(ontologyRLTable, bson.M{"name": rl1OntologyName}) // 这儿找到420条数据

	if KG.KGDataSourceConfig.IsRandom == "yes" {
		authorRl = returnNumSlice(nu, authorRl)
	}

	for _, a := range authorRl {
		// 根据作者属性id 找第一个三元组关系
		da, _ := KG.KGCaseGetterMongo.MongoFind(entityRLTable, bson.M{"ot_rl_id": a.Map()["_id"]}) // 这儿找到一堆的三元组关系1

		if KG.KGDataSourceConfig.IsRandom == "yes" {
			da = returnNumSlice(nu, da)
		}

		for _, b := range da {
			var wg sync.WaitGroup
			var f, q []*bson.D
			wg.Add(1)
			go func() {
				f, _ = KG.KGCaseGetterMongo.MongoFind(entityRLTable, bson.M{"e_id": b.Map()["e_id2"]})
				wg.Done()
			}()
			wg.Add(1)
			go func() {
				// 根据三元组关系找实体1 组装query
				q, _ = KG.KGCaseGetterMongo.MongoFind(entityTable, bson.M{"_id": b.Map()["e_id"], "need_audit": false}, options.Find().SetLimit(1))
				wg.Done()
			}()
			wg.Wait()

			if KG.KGDataSourceConfig.IsRandom == "yes" {
				f = returnNumSlice(nu, f)
			}

			//	根据eid2 找第二个三元组关系
			for _, x := range f {
				var kk, n []*bson.D
				wg.Add(1)
				go func() {
					kk, _ = KG.KGCaseGetterMongo.MongoFind(ontologyRLTable, bson.M{"_id": x.Map()["ot_rl_id"]})
					wg.Done()
				}()
				wg.Add(1)
				go func() {
					n, _ = KG.KGCaseGetterMongo.MongoFind(entityTable, bson.M{"_id": x.Map()["e_id2"], "need_audit": false}, options.Find().SetLimit(1))
					wg.Done()
				}()
				wg.Wait()
				if kk != nil && n != nil && q != nil {
					if kk[0].Map()["name"] == rl2OntologyName {
						Req = &KGTaskReq{
							Query:        mongo.GetInterfaceToString(q[0].Map()["name"]) + model,
							ExpectAnswer: mongo.GetInterfaceToString(n[0].Map()["name"]),
						}
						return
					}
				}
			}
		}
	}
	return nil

}

func (KG *KGTask) getOneTemplate(tmpList []*template) (string, string, string) {
	// 从模板列表中抽取一条出来
	var mu sync.Mutex
	mu.Lock()
	if KG.KGDataSourceConfig.IsRandom == "yes" {
		rand.Seed(time.Now().UnixNano())
		tmp := tmpList[rand.Intn(len(tmpList))]
		model := tmp.Model[rand.Intn(len(tmp.Model))]
		mu.Unlock()
		return tmp.Rl1OntologyName, model.Rl2OntologyName, model.Query
	}
	mu.Unlock()
	// 不随机 就先返回第一个 后面再看下怎么去处理
	return tmpList[0].Rl1OntologyName, tmpList[0].Model[0].Rl2OntologyName, tmpList[0].Model[0].Query
}

// 两跳 模板JSON文件结构
type template struct {
	Rl1OntologyName string `json:"rl1_ontology_name"`
	Model           []struct {
		Query           string `json:"query"`
		Rl2OntologyName string `json:"rl2_ontology_name"`
	} `json:"model"`
}

// 两跳 对查询到的数据组进行切片处理
func returnNumSlice(n int, x []*bson.D) []*bson.D {
	if len(x) > n {
		rand.Seed(time.Now().UnixNano())
		q := rand.Intn(len(x) - n)
		x = x[q : q+n]
	}
	return x
}
