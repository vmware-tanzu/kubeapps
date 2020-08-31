/*
Copyright (c) 2019 Bitnami

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"math"

	"github.com/globalsign/mgo/bson"
	"github.com/kubeapps/common/datastore"
	"github.com/kubeapps/kubeapps/pkg/chart/models"
	"github.com/kubeapps/kubeapps/pkg/dbutils"
)

type mongodbAssetManager struct {
	*dbutils.MongodbAssetManager
}

func newMongoDBManager(config datastore.Config, kubeappsNamespace string) assetManager {
	m := dbutils.NewMongoDBManager(config, kubeappsNamespace)
	return &mongodbAssetManager{m}
}

func (m *mongodbAssetManager) getPaginatedChartList(namespace, repo string, pageNumber, pageSize int, showDuplicates bool) ([]*models.Chart, int, error) {
	db, closer := m.DBSession.DB()
	defer closer()
	var charts []*models.Chart

	c := db.C(chartCollection)
	pipeline := []bson.M{}
	matcher := bson.M{}
	if namespace != dbutils.AllNamespaces {
		matcher["repo.namespace"] = bson.M{"$in": []string{namespace, m.KubeappsNamespace}}
	}
	if repo != "" {
		matcher["repo.name"] = repo
	}
	if len(matcher) > 0 {
		pipeline = append(pipeline, bson.M{"$match": matcher})
	}

	if !showDuplicates {
		// We should query unique charts
		pipeline = append(pipeline,
			// Add a new field to store the latest version
			bson.M{"$addFields": bson.M{"firstChartVersion": bson.M{"$arrayElemAt": []interface{}{"$chartversions", 0}}}},
			// Group by unique digest for the latest version (remove duplicates)
			bson.M{"$group": bson.M{"_id": "$firstChartVersion.digest", "chart": bson.M{"$first": "$$ROOT"}}},
			// Restore original object struct
			bson.M{"$replaceRoot": bson.M{"newRoot": "$chart"}},
		)
	}

	// Order by name
	pipeline = append(pipeline, bson.M{"$sort": bson.M{"name": 1}})

	totalPages := 1
	if pageSize != 0 {
		// If a pageSize is given, returns only the the specified number of charts and
		// the number of pages
		countPipeline := append(pipeline, bson.M{"$count": "count"})
		cc := count{}
		err := c.Pipe(countPipeline).One(&cc)
		if err != nil {
			return charts, 0, err
		}
		totalPages = int(math.Ceil(float64(cc.Count) / float64(pageSize)))

		// If the page number is out of range, return the last one
		if pageNumber > totalPages {
			pageNumber = totalPages
		}

		pipeline = append(pipeline,
			bson.M{"$skip": pageSize * (pageNumber - 1)},
			bson.M{"$limit": pageSize},
		)
	}
	err := c.Pipe(pipeline).All(&charts)
	if err != nil {
		return charts, 0, err
	}

	return charts, totalPages, nil
}

func (m *mongodbAssetManager) getChart(namespace, chartID string) (models.Chart, error) {
	db, closer := m.DBSession.DB()
	defer closer()
	var chart models.Chart
	err := db.C(chartCollection).Find(bson.M{"repo.namespace": namespace, "chart_id": chartID}).One(&chart)
	return chart, err
}

func (m *mongodbAssetManager) getChartVersion(namespace, chartID, version string) (models.Chart, error) {
	db, closer := m.DBSession.DB()
	defer closer()
	var chart models.Chart
	err := db.C(chartCollection).Find(bson.M{
		"repo.namespace": namespace,
		"chart_id":       chartID,
		"chartversions":  bson.M{"$elemMatch": bson.M{"version": version}},
	}).Select(bson.M{
		"name": 1, "repo": 1, "description": 1, "home": 1, "keywords": 1, "maintainers": 1, "sources": 1,
		"chartversions.$": 1,
	}).One(&chart)
	return chart, err
}

func (m *mongodbAssetManager) getChartFiles(namespace, filesID string) (models.ChartFiles, error) {
	db, closer := m.DBSession.DB()
	defer closer()
	var files models.ChartFiles
	err := db.C(filesCollection).Find(bson.M{"repo.namespace": namespace, "file_id": filesID}).One(&files)
	return files, err
}

func (m *mongodbAssetManager) getChartsWithFilters(namespace, name, version, appVersion string) ([]*models.Chart, error) {
	db, closer := m.DBSession.DB()
	defer closer()
	var charts []*models.Chart
	matcher := bson.M{
		"repo.namespace": namespace,
		"name":           name,
		"chartversions": bson.M{
			"$elemMatch": bson.M{"version": version, "appversion": appVersion},
		},
	}
	if namespace != dbutils.AllNamespaces {
		matcher["repo.namespace"] = bson.M{"$in": []string{namespace, m.KubeappsNamespace}}
	}

	err := db.C(chartCollection).Find(matcher).Select(bson.M{
		"name": 1, "repo": 1,
		"chartversions": bson.M{"$slice": 1},
	}).All(&charts)
	return charts, err
}

func (m *mongodbAssetManager) searchCharts(query, repo string) ([]*models.Chart, error) {
	db, closer := m.DBSession.DB()
	defer closer()
	var charts []*models.Chart
	conditions := bson.M{
		"$or": []bson.M{
			{"name": bson.M{"$regex": query}},
			{"description": bson.M{"$regex": query}},
			{"repo.name": bson.M{"$regex": query}},
			{"keywords": bson.M{"$elemMatch": bson.M{"$regex": query}}},
			{"sources": bson.M{"$elemMatch": bson.M{"$regex": query}}},
			{"maintainers": bson.M{"$elemMatch": bson.M{"name": bson.M{"$regex": query}}}},
		},
	}
	if repo != "" {
		conditions["repo.name"] = repo
	}
	err := db.C(chartCollection).Find(conditions).All(&charts)
	return charts, err
}
