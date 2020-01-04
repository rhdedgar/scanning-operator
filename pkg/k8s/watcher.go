package k8s

import (
	managedv1alpha1 "github.com/rhdedgar/scanning-operator/pkg/apis/managed/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// WatcherDaemonSet returns a new daemonset customized for watcher
func WatcherDaemonSet(m *managedv1alpha1.Watcher) *appsv1.DaemonSet {
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
					"name": "watcher",
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"name": "watcher",
					},
				},
				Spec: corev1.PodSpec{
					NodeSelector: map[string]string{
						"beta.kubernetes.io/os": "linux",
					},
					ServiceAccountName: "scanning-operator",
					Tolerations: []corev1.Toleration{
						{
							Operator: corev1.TolerationOpExists,
						},
					},
					Containers: []corev1.Container{{
						Image: "quay.io/dedgar/pleg-watcher:latest",
						Name:  "watcher",
						SecurityContext: &corev1.SecurityContext{
							Privileged: &privileged,
							RunAsUser:  &runAsUser,
						},
						// TODO/dedgar consider pulling env var defaults from a config pkg or CR.
						Env: []corev1.EnvVar{{
							Name:  "OO_PAUSE_ON_START",
							Value: "false",
						}, {
							Name:  "CRIO_LOG_URL",
							Value: "http://logger.openshift-scanning-operator.svc:8080/api/crio/log",
						}, {
							Name:  "DOCKER_LOG_URL",
							Value: "http://logger.openshift-scanning-operator.svc:8080/api/docker/log",
						}, {
							Name:  "JOURNAL_PATH",
							Value: "/var/log/journal",
						}, {
							Name:  "SCAN_RESULTS_DIR",
							Value: "",
						}, {
							Name:  "POST_RESULT_URL",
							Value: "http://logger.openshift-scanning-operator.svc:8080/api/clam/scanresult",
						}, {
							Name:  "OUT_FILE",
							Value: "",
						}, {
							Name:  "CLAM_SOCKET",
							Value: "/host/var/run/clamd.scan/clamd.sock",
						}},
						Resources: corev1.ResourceRequirements{},
						VolumeMounts: []corev1.VolumeMount{{
							Name:      "watcher-host-journal",
							MountPath: "/var/log/journal",
						}, {
							Name:      "watcher-host-filesystem",
							MountPath: "/host/",
						}},
					}},
					Volumes: []corev1.Volume{{
						Name: "watcher-host-filesystem",
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
