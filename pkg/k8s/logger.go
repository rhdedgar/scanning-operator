package k8s

import (
	managedv1alpha1 "github.com/rhdedgar/scanning-operator/pkg/apis/managed/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// LoggerDaemonSet returns a new daemonset customized for logger
func LoggerDaemonSet(m *managedv1alpha1.Logger) *appsv1.DaemonSet {
	ds := &appsv1.DaemonSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      m.Name,
			Namespace: m.Namespace,
		},
		Spec: appsv1.DaemonSetSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"name": "logger",
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"name": "logger",
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
						Image: "quay.io/dedgar/pod-logger:latest",
						Name:  "logger",
						// TODO/dedgar consider pulling env var defaults from a config pkg or CR.
						Env: []corev1.EnvVar{{
							Name:  "OO_PAUSE_ON_START",
							Value: "false",
						}, {
							Name:  "LOG_WRITER_URL",
							Value: "http://logger.openshift-scanning-operator.svc:8080/api/log",
						}},
						Ports: []corev1.ContainerPort{{
							ContainerPort: 8080,
							Name:          "logger",
						}},
						Resources: corev1.ResourceRequirements{},
						VolumeMounts: []corev1.VolumeMount{{
							Name:      "logger-secrets",
							MountPath: "/secrets",
						}},
					}},
					Volumes: []corev1.Volume{{
						Name: "logger-secrets",
						VolumeSource: corev1.VolumeSource{
							Secret: &corev1.SecretVolumeSource{
								SecretName: "logger-secrets",
							},
						},
					}},
				},
			},
		},
	}
	return ds
}
