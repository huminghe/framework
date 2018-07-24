package orm

import (
	"context"
	log "github.com/cihub/seelog"
	"github.com/golang/go/src/pkg/fmt"
	"github.com/infinitbyte/framework/core/errors"
	"github.com/infinitbyte/framework/core/index"
	api "github.com/infinitbyte/framework/core/persist"
	"github.com/infinitbyte/framework/core/util"
	"github.com/olivere/elastic"
)

type ElasticORM struct {
	Client *index.ElasticsearchClient
}

func getIndex(any interface{}) string {
	return util.GetTypeName(any, true)
}

func getID(any interface{}) string {
	return util.GetFieldValueByTagName(any, "index", "id")
}

func (handler ElasticORM) RegisterSchema(t interface{}) error {
	// TODO create index and generate mapping
	return nil
}

func (handler ElasticORM) Get(o interface{}) error {

	response, err := handler.Client.Get(getIndex(o), getID(o))
	if err != nil {
		return err
	}

	//TODO improve performance
	str := util.ToJson(response.Source, false)
	return util.FromJson(str, o)
}

func (handler ElasticORM) GetBy(field string, value interface{}, t interface{}, to interface{}) (error, api.Result) {

	query := api.Query{}
	query.Conds = api.And(api.Eq(field, value))
	return handler.Search(t, to, &query)
}

func (handler ElasticORM) Save(o interface{}) error {
	_, err := handler.Client.Index(getIndex(o), getID(o), o)
	return err
}

func (handler ElasticORM) Update(o interface{}) error {
	return handler.Save(o)
}

func (handler ElasticORM) Delete(o interface{}) error {
	_, err := handler.Client.Delete(getIndex(o), getID(o))
	return err
}

func (handler ElasticORM) Count(o interface{}) (int, error) {
	countResponse, err := handler.Client.Count(getIndex(o))
	if err != nil {
		return 0, err
	}
	return countResponse.Count, err
}

func getQuery(c1 *api.Cond) interface{} {

	switch c1.QueryType {
	case api.Match:
		q := index.MatchQuery{}
		q.Set(c1.Field, c1.Value)
		return q
	case api.RangeGt:
		q := index.RangeQuery{}
		q.Gt(c1.Field, c1.Value)
		return q
	case api.RangeGte:
		q := index.RangeQuery{}
		q.Gte(c1.Field, c1.Value)
		return q
	case api.RangeLt:
		q := index.RangeQuery{}
		q.Lt(c1.Field, c1.Value)
		return q
	case api.RangeLte:
		q := index.RangeQuery{}
		q.Lte(c1.Field, c1.Value)
		return q
	}
	panic(errors.Errorf("invalid query: %s", c1))
}

func (handler ElasticORM) Search(t interface{}, to interface{}, q *api.Query) (error, api.Result) {

	var err error

	request := index.SearchRequest{}

	request.From = q.From
	request.Size = q.Size

	if q.Conds != nil && len(q.Conds) > 0 {
		request.Query = &index.Query{}

		boolQuery := index.BoolQuery{}

		for _, c1 := range q.Conds {
			q := getQuery(c1)
			switch c1.BoolType {
			case api.Must:
				boolQuery.Must = append(boolQuery.Must, q)
				break
			case api.MustNot:
				boolQuery.MustNot = append(boolQuery.MustNot, q)
				break
			case api.Should:
				boolQuery.Should = append(boolQuery.Should, q)
				break
			}

		}

		request.Query.BoolQuery = &boolQuery

	}

	if q.Sort != nil && len(*q.Sort) > 0 {
		for _, i := range *q.Sort {
			request.AddSort(i.Field, string(i.SortType))
		}
	}

	result := api.Result{}
	searchResponse, err := handler.Client.Search(getIndex(t), &request)
	if err != nil {
		return err, result
	}

	var array []interface{}

	for _, doc := range searchResponse.Hits.Hits {
		array = append(array, doc.Source)
	}

	result.Result = array
	result.Total = searchResponse.Hits.Total

	return err, result
}

func (handler ElasticORM) GroupBy(t interface{}, selectField, groupField string, haveQuery string, haveValue interface{}) (error, map[string]interface{}) {

	agg := elastic.NewTermsAggregation().Field(selectField).Size(10)

	// Create an Elasticsearch client
	client, err := elastic.NewClient(elastic.SetURL(handler.Client.Config.Endpoint), elastic.SetSniff(true))
	if err != nil {
		log.Error(err)
	}

	indexName := getIndex(t)
	if handler.Client.Config.IndexPrefix != "" {
		indexName = handler.Client.Config.IndexPrefix + indexName
	}

	result, err := client.Search(indexName).Aggregation(selectField, agg).Do(context.TODO())
	if err != nil {
		log.Error(err)
	}

	finalResult := map[string]interface{}{}

	items, ok := result.Aggregations.Terms(selectField)
	if ok {
		log.Error(selectField)
		for _, item := range items.Buckets {
			k := fmt.Sprintf("%v", item.Key)
			finalResult[k] = item.DocCount
			log.Trace(item.Key, ":", item.DocCount)
		}
	}

	//panic(errors.New("not implemented yet"))
	return nil, finalResult
}