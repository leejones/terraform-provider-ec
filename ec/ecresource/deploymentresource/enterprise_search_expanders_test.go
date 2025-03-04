// Licensed to Elasticsearch B.V. under one or more contributor
// license agreements. See the NOTICE file distributed with
// this work for additional information regarding copyright
// ownership. Elasticsearch B.V. licenses this file to you under
// the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

package deploymentresource

import (
	"errors"
	"testing"

	"github.com/elastic/cloud-sdk-go/pkg/api/mock"
	"github.com/elastic/cloud-sdk-go/pkg/models"
	"github.com/elastic/cloud-sdk-go/pkg/util/ec"
	"github.com/stretchr/testify/assert"
)

func Test_expandEssResources(t *testing.T) {
	tpl := func() *models.EnterpriseSearchPayload {
		return essResource(parseDeploymentTemplate(t,
			"testdata/template-aws-io-optimized-v2.json",
		))
	}
	type args struct {
		ess []interface{}
		tpl *models.EnterpriseSearchPayload
	}
	tests := []struct {
		name string
		args args
		want []*models.EnterpriseSearchPayload
		err  error
	}{
		{
			name: "returns nil when there's no resources",
		},
		{
			name: "parses an enterprise_search resource with explicit topology",
			args: args{
				tpl: tpl(),
				ess: []interface{}{map[string]interface{}{
					"ref_id":                       "main-enterprise_search",
					"resource_id":                  mock.ValidClusterID,
					"version":                      "7.7.0",
					"region":                       "some-region",
					"elasticsearch_cluster_ref_id": "somerefid",
					"topology": []interface{}{map[string]interface{}{
						"instance_configuration_id": "aws.enterprisesearch.m5d",
						"size":                      "2g",
						"zone_count":                1,
					}},
				}},
			},
			want: []*models.EnterpriseSearchPayload{{
				ElasticsearchClusterRefID: ec.String("somerefid"),
				Region:                    ec.String("some-region"),
				RefID:                     ec.String("main-enterprise_search"),
				Plan: &models.EnterpriseSearchPlan{
					EnterpriseSearch: &models.EnterpriseSearchConfiguration{
						Version: "7.7.0",
					},
					ClusterTopology: []*models.EnterpriseSearchTopologyElement{{
						ZoneCount:               1,
						InstanceConfigurationID: "aws.enterprisesearch.m5d",
						Size: &models.TopologySize{
							Resource: ec.String("memory"),
							Value:    ec.Int32(2048),
						},
						NodeType: &models.EnterpriseSearchNodeTypes{
							Appserver: ec.Bool(true),
							Connector: ec.Bool(true),
							Worker:    ec.Bool(true),
						},
					}},
				},
			}},
		},
		{
			name: "parses an enterprise_search resource with no topology takes the minimum size",
			args: args{
				tpl: tpl(),
				ess: []interface{}{map[string]interface{}{
					"ref_id":                       "main-enterprise_search",
					"resource_id":                  mock.ValidClusterID,
					"version":                      "7.7.0",
					"region":                       "some-region",
					"elasticsearch_cluster_ref_id": "somerefid",
				}},
			},
			want: []*models.EnterpriseSearchPayload{{
				ElasticsearchClusterRefID: ec.String("somerefid"),
				Region:                    ec.String("some-region"),
				RefID:                     ec.String("main-enterprise_search"),
				Plan: &models.EnterpriseSearchPlan{
					EnterpriseSearch: &models.EnterpriseSearchConfiguration{
						Version: "7.7.0",
					},
					ClusterTopology: []*models.EnterpriseSearchTopologyElement{{
						ZoneCount:               2,
						InstanceConfigurationID: "aws.enterprisesearch.m5d",
						Size: &models.TopologySize{
							Resource: ec.String("memory"),
							Value:    ec.Int32(2048),
						},
						NodeType: &models.EnterpriseSearchNodeTypes{
							Appserver: ec.Bool(true),
							Connector: ec.Bool(true),
							Worker:    ec.Bool(true),
						},
					}},
				},
			}},
		},
		{
			name: "parses an enterprise_search resource with topology but no instance_configuration_id",
			args: args{
				tpl: tpl(),
				ess: []interface{}{map[string]interface{}{
					"ref_id":                       "main-enterprise_search",
					"resource_id":                  mock.ValidClusterID,
					"version":                      "7.7.0",
					"region":                       "some-region",
					"elasticsearch_cluster_ref_id": "somerefid",
					"topology": []interface{}{map[string]interface{}{
						"size": "4g",
					}},
				}},
			},
			want: []*models.EnterpriseSearchPayload{{
				ElasticsearchClusterRefID: ec.String("somerefid"),
				Region:                    ec.String("some-region"),
				RefID:                     ec.String("main-enterprise_search"),
				Plan: &models.EnterpriseSearchPlan{
					EnterpriseSearch: &models.EnterpriseSearchConfiguration{
						Version: "7.7.0",
					},
					ClusterTopology: []*models.EnterpriseSearchTopologyElement{{
						ZoneCount:               2,
						InstanceConfigurationID: "aws.enterprisesearch.m5d",
						Size: &models.TopologySize{
							Resource: ec.String("memory"),
							Value:    ec.Int32(4096),
						},
						NodeType: &models.EnterpriseSearchNodeTypes{
							Appserver: ec.Bool(true),
							Connector: ec.Bool(true),
							Worker:    ec.Bool(true),
						},
					}},
				},
			}},
		},
		{
			name: "parses an enterprise_search resource with multiple topologies but no instance_configuration_id",
			args: args{
				tpl: tpl(),
				ess: []interface{}{map[string]interface{}{
					"ref_id":                       "main-enterprise_search",
					"resource_id":                  mock.ValidClusterID,
					"version":                      "7.7.0",
					"region":                       "some-region",
					"elasticsearch_cluster_ref_id": "somerefid",
					"topology": []interface{}{
						map[string]interface{}{
							"size": "4g",
						}, map[string]interface{}{
							"size": "4g",
						},
					},
				}},
			},
			err: errors.New("enterprise_search topology: invalid instance_configuration_id: \"\" doesn't match any of the deployment template instance configurations"),
		},
		{
			name: "parses an enterprise_search resource with topology but instance_configuration_id",
			args: args{
				tpl: tpl(),
				ess: []interface{}{map[string]interface{}{
					"ref_id":                       "main-enterprise_search",
					"resource_id":                  mock.ValidClusterID,
					"version":                      "7.7.0",
					"region":                       "some-region",
					"elasticsearch_cluster_ref_id": "somerefid",
					"topology": []interface{}{map[string]interface{}{
						"instance_configuration_id": "aws.enterprisesearch.m5d",
					}},
				}},
			},
			want: []*models.EnterpriseSearchPayload{{
				ElasticsearchClusterRefID: ec.String("somerefid"),
				Region:                    ec.String("some-region"),
				RefID:                     ec.String("main-enterprise_search"),
				Plan: &models.EnterpriseSearchPlan{
					EnterpriseSearch: &models.EnterpriseSearchConfiguration{
						Version: "7.7.0",
					},
					ClusterTopology: []*models.EnterpriseSearchTopologyElement{{
						ZoneCount:               2,
						InstanceConfigurationID: "aws.enterprisesearch.m5d",
						Size: &models.TopologySize{
							Resource: ec.String("memory"),
							Value:    ec.Int32(2048),
						},
						NodeType: &models.EnterpriseSearchNodeTypes{
							Appserver: ec.Bool(true),
							Connector: ec.Bool(true),
							Worker:    ec.Bool(true),
						},
					}},
				},
			}},
		},
		{
			name: "parses an enterprise_search resource with topology and zone_count",
			args: args{
				tpl: tpl(),
				ess: []interface{}{map[string]interface{}{
					"ref_id":                       "main-enterprise_search",
					"resource_id":                  mock.ValidClusterID,
					"version":                      "7.7.0",
					"region":                       "some-region",
					"elasticsearch_cluster_ref_id": "somerefid",
					"topology": []interface{}{map[string]interface{}{
						"zone_count": 1,
					}},
				}},
			},
			want: []*models.EnterpriseSearchPayload{{
				ElasticsearchClusterRefID: ec.String("somerefid"),
				Region:                    ec.String("some-region"),
				RefID:                     ec.String("main-enterprise_search"),
				Plan: &models.EnterpriseSearchPlan{
					EnterpriseSearch: &models.EnterpriseSearchConfiguration{
						Version: "7.7.0",
					},
					ClusterTopology: []*models.EnterpriseSearchTopologyElement{{
						ZoneCount:               1,
						InstanceConfigurationID: "aws.enterprisesearch.m5d",
						Size: &models.TopologySize{
							Resource: ec.String("memory"),
							Value:    ec.Int32(2048),
						},
						NodeType: &models.EnterpriseSearchNodeTypes{
							Appserver: ec.Bool(true),
							Connector: ec.Bool(true),
							Worker:    ec.Bool(true),
						},
					}},
				},
			}},
		},
		{
			name: "parses an enterprise_search resource with explicit topology and config",
			args: args{
				tpl: tpl(),
				ess: []interface{}{map[string]interface{}{
					"ref_id":                       "secondary-enterprise_search",
					"elasticsearch_cluster_ref_id": "somerefid",
					"resource_id":                  mock.ValidClusterID,
					"version":                      "7.7.0",
					"region":                       "some-region",
					"config": []interface{}{map[string]interface{}{
						"user_settings_yaml":          "some.setting: value",
						"user_settings_override_yaml": "some.setting: override",
						"user_settings_json":          `{"some.setting":"value"}`,
						"user_settings_override_json": `{"some.setting":"override"}`,
					}},
					"topology": []interface{}{map[string]interface{}{
						"instance_configuration_id": "aws.enterprisesearch.m5d",
						"size":                      "4g",
						"zone_count":                1,
						"node_type_appserver":       true,
						"node_type_connector":       true,
						"node_type_worker":          true,
					}},
				}},
			},
			want: []*models.EnterpriseSearchPayload{{
				ElasticsearchClusterRefID: ec.String("somerefid"),
				Region:                    ec.String("some-region"),
				RefID:                     ec.String("secondary-enterprise_search"),
				Plan: &models.EnterpriseSearchPlan{
					EnterpriseSearch: &models.EnterpriseSearchConfiguration{
						Version:                  "7.7.0",
						UserSettingsYaml:         "some.setting: value",
						UserSettingsOverrideYaml: "some.setting: override",
						UserSettingsJSON: map[string]interface{}{
							"some.setting": "value",
						},
						UserSettingsOverrideJSON: map[string]interface{}{
							"some.setting": "override",
						},
					},
					ClusterTopology: []*models.EnterpriseSearchTopologyElement{{
						ZoneCount:               1,
						InstanceConfigurationID: "aws.enterprisesearch.m5d",
						Size: &models.TopologySize{
							Resource: ec.String("memory"),
							Value:    ec.Int32(4096),
						},
						NodeType: &models.EnterpriseSearchNodeTypes{
							Appserver: ec.Bool(true),
							Connector: ec.Bool(true),
							Worker:    ec.Bool(true),
						},
					}},
				},
			}},
		},
		{
			name: "parses an enterprise_search resource with explicit nils",
			args: args{
				tpl: tpl(),
				ess: []interface{}{map[string]interface{}{
					"ref_id":                       "secondary-enterprise_search",
					"elasticsearch_cluster_ref_id": "somerefid",
					"resource_id":                  mock.ValidClusterID,
					"version":                      "7.7.0",
					"region":                       nil,
					"config": []interface{}{map[string]interface{}{
						"user_settings_yaml":          nil,
						"user_settings_override_yaml": nil,
						"user_settings_json":          nil,
						"user_settings_override_json": nil,
					}},
					"topology": nil,
				}},
			},
			want: []*models.EnterpriseSearchPayload{{
				ElasticsearchClusterRefID: ec.String("somerefid"),
				Region:                    ec.String("us-east-1"),
				RefID:                     ec.String("secondary-enterprise_search"),
				Plan: &models.EnterpriseSearchPlan{
					EnterpriseSearch: &models.EnterpriseSearchConfiguration{
						Version: "7.7.0",
					},
					ClusterTopology: []*models.EnterpriseSearchTopologyElement{{
						ZoneCount:               2,
						InstanceConfigurationID: "aws.enterprisesearch.m5d",
						Size: &models.TopologySize{
							Resource: ec.String("memory"),
							Value:    ec.Int32(2048),
						},
						NodeType: &models.EnterpriseSearchNodeTypes{
							Appserver: ec.Bool(true),
							Connector: ec.Bool(true),
							Worker:    ec.Bool(true),
						},
					}},
				},
			}},
		},
		{
			name: "parses an enterprise_search resource with invalid instance_configuration_id",
			args: args{
				tpl: tpl(),
				ess: []interface{}{map[string]interface{}{
					"ref_id":                       "main-enterprise_search",
					"resource_id":                  mock.ValidClusterID,
					"version":                      "7.7.0",
					"region":                       "some-region",
					"elasticsearch_cluster_ref_id": "somerefid",
					"topology": []interface{}{map[string]interface{}{
						"instance_configuration_id": "aws.enterprisesearch.m5",
						"size":                      "2g",
						"zone_count":                1,
					}},
				}},
			},
			err: errors.New(`enterprise_search topology: invalid instance_configuration_id: "aws.enterprisesearch.m5" doesn't match any of the deployment template instance configurations`),
		},
		{
			name: "tries to parse an enterprise_search resource when the template doesn't have an Enterprise Search instance set.",
			args: args{
				tpl: nil,
				ess: []interface{}{map[string]interface{}{
					"ref_id":                       "tertiary-enterprise_search",
					"elasticsearch_cluster_ref_id": "somerefid",
					"resource_id":                  mock.ValidClusterID,
					"version":                      "7.8.0",
					"region":                       "some-region",
					"config": []interface{}{map[string]interface{}{
						"user_settings_yaml":          "some.setting: value",
						"user_settings_override_yaml": "some.setting: value2",
						"user_settings_json":          "{\"some.setting\": \"value\"}",
						"user_settings_override_json": "{\"some.setting\": \"value2\"}",
					}},
					"topology": []interface{}{map[string]interface{}{
						"instance_configuration_id": "aws.enterprise_search.r5d",
						"size":                      "4g",
						"size_resource":             "memory",
						"zone_count":                1,
					}},
				}},
			},
			err: errors.New("enterprise_search specified but deployment template is not configured for it. Use a different template if you wish to add enterprise_search"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := expandEssResources(tt.args.ess, tt.args.tpl)
			if !assert.Equal(t, tt.err, err) {
				t.Error(err)
			}

			assert.Equal(t, tt.want, got)
		})
	}
}
