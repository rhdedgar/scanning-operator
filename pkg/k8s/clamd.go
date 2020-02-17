package k8s

import (
	managedv1alpha1 "github.com/rhdedgar/scanning-operator/pkg/apis/managed/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ClamdDaemonSet returns a new daemonset customized for clamd
func ClamdDaemonSet(m *managedv1alpha1.Clamd) *appsv1.DaemonSet {
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
					"name": "clamd",
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"name": "clamd",
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
					InitContainers: []corev1.Container{{
						Image:     "quay.io/dedgar/clamsig-puller:latest",
						Name:      "init-clamsig-puller",
						Resources: corev1.ResourceRequirements{},
						Env: []corev1.EnvVar{{
							Name:  "OO_PAUSE_ON_START",
							Value: "false",
						}, {
							Name:  "INIT_CONTAINER",
							Value: "true",
						}},
						VolumeMounts: []corev1.VolumeMount{{
							Name:      "clam-secrets",
							MountPath: "/secrets",
						}, {
							Name:      "clam-files",
							MountPath: "/clam",
						}},
					}},
					Containers: []corev1.Container{{
						Image:     "quay.io/dedgar/clamsig-puller:latest",
						Name:      "clamsig-puller",
						Resources: corev1.ResourceRequirements{},
						Env: []corev1.EnvVar{{
							Name:  "OO_PAUSE_ON_START",
							Value: "false",
						}},
						VolumeMounts: []corev1.VolumeMount{{
							Name:      "clam-secrets",
							MountPath: "/secrets",
						}, {
							Name:      "clam-files",
							MountPath: "/clam",
						}},
					}, {
						Image:     "quay.io/dedgar/clamd:latest",
						Name:      "clamd",
						Resources: corev1.ResourceRequirements{},
						Env: []corev1.EnvVar{{
							Name:  "OO_PAUSE_ON_START",
							Value: "false",
						}},
						VolumeMounts: []corev1.VolumeMount{{
							Name:      "clam-files",
							MountPath: "/var/lib/clamav",
						}},
					}, {
						Image: "quay.io/dedgar/watcher:latest",
						Name:  "watcher",
						SecurityContext: &corev1.SecurityContext{
							Privileged: &privileged,
							RunAsUser:  &runAsUser,
						},
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
							Name:  "CHROOT_PATH",
							Value: "/host",
						}, {
							Name:  "CLAM_SOCKET",
							Value: "/clam/clamd.sock",
						}},
						Resources: corev1.ResourceRequirements{},
						VolumeMounts: []corev1.VolumeMount{{
							Name:      "watcher-host-journal",
							MountPath: "/var/log/journal",
						}, {
							Name:      "watcher-host-filesystem",
							MountPath: "/host",
						}, {
							Name:      "clam-files",
							MountPath: "/clam",
						}},
					}},
					Volumes: []corev1.Volume{{
						Name: "clam-secrets",
						VolumeSource: corev1.VolumeSource{
							Secret: &corev1.SecretVolumeSource{
								SecretName: "clam-secrets",
							},
						},
					}, {
						Name: "clam-files",
						VolumeSource: corev1.VolumeSource{
							EmptyDir: &corev1.EmptyDirVolumeSource{},
						},
					}, {
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
