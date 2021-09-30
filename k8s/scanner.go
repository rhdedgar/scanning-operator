/*
Copyright 2020 Doug Edgar.

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

package k8s

import (
	managedv1alpha1 "github.com/rhdedgar/scanning-operator/api/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ScannerDaemonSet returns a new daemonset customized for scanner
func ScannerDaemonSet(m *managedv1alpha1.Scanner) *appsv1.DaemonSet {
	var privileged = true
	var runAsUser int64

	ds := &appsv1.DaemonSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      m.Name,
			Namespace: m.Namespace,
		},
		Spec: appsv1.DaemonSetSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"name": "scanner",
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"name": "scanner",
					},
				},
				Spec: corev1.PodSpec{
					NodeSelector: map[string]string{
						"beta.kubernetes.io/os": "linux",
					},
					// ServiceAccountName: "openshift-scanning-operator",
					Tolerations: []corev1.Toleration{
						{
							Operator: corev1.TolerationOpExists,
						},
					},
					InitContainers: []corev1.Container{{
						Image: "quay.io/dedgar/clamsig-puller:v0.0.4",
						Name:  "init-clamsig-puller",
						Resources: corev1.ResourceRequirements{
							Requests: corev1.ResourceList{
								corev1.ResourceCPU:    resource.MustParse("100m"),
								corev1.ResourceMemory: resource.MustParse("50Mi"),
							},
							Limits: corev1.ResourceList{
								corev1.ResourceCPU:    resource.MustParse("100m"),
								corev1.ResourceMemory: resource.MustParse("200Mi"),
							},
						},
						Env: []corev1.EnvVar{{
							Name:  "OO_PAUSE_ON_START",
							Value: "false",
						}, {
							Name:  "INIT_CONTAINER",
							Value: "true",
						}},
						VolumeMounts: []corev1.VolumeMount{{
							Name:      "scanner-secrets",
							MountPath: "/secrets",
						}, {
							Name:      "clam-files",
							MountPath: "/clam",
						}},
					}},
					Containers: []corev1.Container{{
						Image: "quay.io/dedgar/clamsig-puller:v0.0.4",
						Name:  "clamsig-puller",
						Resources: corev1.ResourceRequirements{
							Requests: corev1.ResourceList{
								corev1.ResourceCPU:    resource.MustParse("100m"),
								corev1.ResourceMemory: resource.MustParse("50Mi"),
							},
							Limits: corev1.ResourceList{
								corev1.ResourceCPU:    resource.MustParse("100m"),
								corev1.ResourceMemory: resource.MustParse("200Mi"),
							},
						},
						Env: []corev1.EnvVar{{
							Name:  "OO_PAUSE_ON_START",
							Value: "false",
						}},
						VolumeMounts: []corev1.VolumeMount{{
							Name:      "scanner-secrets",
							MountPath: "/secrets",
						}, {
							Name:      "clam-files",
							MountPath: "/clam",
						}},
					}, {
						Image: "quay.io/dedgar/clamd:v0.0.3",
						Name:  "clamd",
						Resources: corev1.ResourceRequirements{
							Requests: corev1.ResourceList{
								corev1.ResourceCPU:    resource.MustParse("100m"),
								corev1.ResourceMemory: resource.MustParse("800Mi"),
							},
							Limits: corev1.ResourceList{
								corev1.ResourceCPU:    resource.MustParse("300m"),
								corev1.ResourceMemory: resource.MustParse("2Gi"),
							},
						},
						Env: []corev1.EnvVar{{
							Name:  "OO_PAUSE_ON_START",
							Value: "false",
						}},
						VolumeMounts: []corev1.VolumeMount{{
							Name:      "clam-files",
							MountPath: "/var/lib/clamav",
						}},
					}, {
						Image: "quay.io/dedgar/container-info:v0.0.15",
						Name:  "info",
						SecurityContext: &corev1.SecurityContext{
							Privileged: &privileged,
							RunAsUser:  &runAsUser,
						},
						Resources: corev1.ResourceRequirements{
							Requests: corev1.ResourceList{
								corev1.ResourceCPU:    resource.MustParse("100m"),
								corev1.ResourceMemory: resource.MustParse("20Mi"),
							},
							Limits: corev1.ResourceList{
								corev1.ResourceCPU:    resource.MustParse("100m"),
								corev1.ResourceMemory: resource.MustParse("50Mi"),
							},
						}, Env: []corev1.EnvVar{{
							Name:  "OO_PAUSE_ON_START",
							Value: "false",
						}, {
							Name:  "CHROOT_PATH",
							Value: "/host",
						}},
						VolumeMounts: []corev1.VolumeMount{{
							Name:      "clam-files",
							MountPath: "/clam",
						}, {
							Name:      "host-filesystem",
							MountPath: "/host",
						}},
					}, {
						Image: "quay.io/dedgar/watcher:v0.0.67",
						Name:  "watcher",
						SecurityContext: &corev1.SecurityContext{
							Privileged: &privileged,
							RunAsUser:  &runAsUser,
						},
						Env: []corev1.EnvVar{{
							Name:  "OO_PAUSE_ON_START",
							Value: "false",
						}, {
							Name:  "ACTIVE_SCAN",
							Value: "true",
						}, {
							Name:  "CRIO_LOG_URL",
							Value: "http://loggerservice.openshift-scanning-operator.svc.cluster.local:8080/api/crio/log",
						}, {
							Name:  "DOCKER_LOG_URL",
							Value: "http://loggerservice.openshift-scanning-operator.svc.cluster.local:8080/api/docker/log",
						}, {
							Name:  "CLAM_LOG_URL",
							Value: "http://loggerservice.openshift-scanning-operator.svc.cluster.local:8080/api/clam/scanresult",
						}, {
							Name:  "JOURNAL_PATH",
							Value: "/var/log/journal",
						}, {
							Name:  "SCAN_RESULTS_DIR",
							Value: "",
						}, {
							Name:  "POST_RESULT_URL",
							Value: "http://loggerservice.openshift-scanning-operator.svc.cluster.local:8080/api/clam/scanresult",
						}, {
							Name:  "OUT_FILE",
							Value: "",
						}, {
							Name:  "SKIP_NAMESPACE_PREFIXES",
							Value: "openshift-",
						}, {
							Name:  "SKIP_NAMESPACES",
							Value: "openshift-scanning-operator,ci",
						}, {
							Name:  "CLAM_SOCKET",
							Value: "/clam/clamd.sock",
						}, {
							Name:  "INFO_SOCKET",
							Value: "@rpc.sock",
						}},
						Resources: corev1.ResourceRequirements{
							Requests: corev1.ResourceList{
								corev1.ResourceCPU:    resource.MustParse("100m"),
								corev1.ResourceMemory: resource.MustParse("50Mi"),
							},
							Limits: corev1.ResourceList{
								corev1.ResourceCPU:    resource.MustParse("100m"),
								corev1.ResourceMemory: resource.MustParse("100Mi"),
							},
						},
						VolumeMounts: []corev1.VolumeMount{{
							Name:      "watcher-host-journal",
							MountPath: "/var/log/journal",
						}, {
							Name:      "host-filesystem",
							MountPath: "/host",
						}, {
							Name:      "clam-files",
							MountPath: "/clam",
						}},
					}, {
						Image: "quay.io/dedgar/watcher:v0.0.67",
						Name:  "scheduler",
						SecurityContext: &corev1.SecurityContext{
							Privileged: &privileged,
							RunAsUser:  &runAsUser,
						},
						Env: []corev1.EnvVar{{
							Name:  "OO_PAUSE_ON_START",
							Value: "false",
						}, {
							Name:  "CRIO_LOG_URL",
							Value: "http://loggerservice.openshift-scanning-operator.svc.cluster.local:8080/api/crio/log",
						}, {
							Name:  "DOCKER_LOG_URL",
							Value: "http://loggerservice.openshift-scanning-operator.svc.cluster.local:8080/api/docker/log",
						}, {
							Name:  "CLAM_LOG_URL",
							Value: "http://loggerservice.openshift-scanning-operator.svc.cluster.local:8080/api/clam/scanresult",
						}, {
							Name:  "JOURNAL_PATH",
							Value: "/var/log/journal",
						}, {
							Name:  "SCAN_RESULTS_DIR",
							Value: "",
						}, {
							Name:  "POST_RESULT_URL",
							Value: "http://loggerservice.openshift-scanning-operator.svc.cluster.local:8080/api/clam/scanresult",
						}, {
							Name:  "OUT_FILE",
							Value: "",
						}, {
							Name:  "CLAM_SOCKET",
							Value: "/clam/clamd.sock",
						}, {
							Name:  "INFO_SOCKET",
							Value: "@rpc.sock",
						}, {
							Name:  "SCHEDULED_SCAN",
							Value: "true",
						}, {
							Name:  "SCHEDULED_SCAN_DAY",
							Value: "Saturday",
						}, {
							Name:  "MIN_CON_DAY",
							Value: "0",
						}, {
							Name:  "SKIP_NAMESPACE_PREFIXES",
							Value: "openshift-",
						}, {
							Name:  "SKIP_NAMESPACES",
							Value: "openshift-scanning-operator",
						}, {
							Name:  "HOST_SCAN_DIRS",
							Value: "/host",
						}},
						Resources: corev1.ResourceRequirements{
							Requests: corev1.ResourceList{
								corev1.ResourceCPU:    resource.MustParse("100m"),
								corev1.ResourceMemory: resource.MustParse("50Mi"),
							},
							Limits: corev1.ResourceList{
								corev1.ResourceCPU:    resource.MustParse("100m"),
								corev1.ResourceMemory: resource.MustParse("100Mi"),
							},
						}, VolumeMounts: []corev1.VolumeMount{{
							Name:      "watcher-host-journal",
							MountPath: "/var/log/journal",
						}, {
							Name:      "host-filesystem",
							MountPath: "/host",
						}, {
							Name:      "clam-files",
							MountPath: "/clam",
						}},
					}},
					Volumes: []corev1.Volume{{
						Name: "scanner-secrets",
						VolumeSource: corev1.VolumeSource{
							Secret: &corev1.SecretVolumeSource{
								SecretName: "scanner-secrets",
							},
						},
					}, {
						Name: "clam-files",
						VolumeSource: corev1.VolumeSource{
							EmptyDir: &corev1.EmptyDirVolumeSource{},
						},
					}, {
						Name: "host-filesystem",
						VolumeSource: corev1.VolumeSource{
							HostPath: &corev1.HostPathVolumeSource{
								Path: "/",
							},
						},
					}, {
						Name: "watcher-host-journal",
						VolumeSource: corev1.VolumeSource{
							HostPath: &corev1.HostPathVolumeSource{
								Path: "/var/log/journal",
							},
						},
					}},
				},
			},
		},
	}
	return ds
}
