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
					Containers: []corev1.Container{{
						Image: "quay.io/dedgar/clam-server:latest",
						Name:  "clamd",
						SecurityContext: &corev1.SecurityContext{
							Privileged: &privileged,
							RunAsUser:  &runAsUser,
						},
						Resources: corev1.ResourceRequirements{},
						VolumeMounts: []corev1.VolumeMount{{
							Name:      "clamd-host-filesystem",
							MountPath: "/host/var/run/clamd.scan",
						}, {
							Name:      "clamd-secrets",
							MountPath: "/secrets",
						}},
					}},
					Volumes: []corev1.Volume{{
						Name: "clamd-host-filesystem",
						VolumeSource: corev1.VolumeSource{
							HostPath: &corev1.HostPathVolumeSource{
								Path: "/host/var/run/clamd.scan",
							},
						},
					}, {
						Name: "clamd-secrets",
						VolumeSource: corev1.VolumeSource{
							Secret: &corev1.SecretVolumeSource{
								SecretName: "clamd-secrets",
							},
						},
					}},
				},
			},
		},
	}
	return ds
}
