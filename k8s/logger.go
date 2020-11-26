package k8s

import (
	managedv1alpha1 "github.com/rhdedgar/scanning-operator/api/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

// LoggerDaemonSet returns a new daemonset customized for logger
func LoggerDaemonSet(m *managedv1alpha1.Logger) *appsv1.DaemonSet {
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
						"node-role.kubernetes.io/master": "",
					},
					ServiceAccountName: "scanning-operator",
					Tolerations: []corev1.Toleration{
						{
							Operator: corev1.TolerationOpExists,
						},
					},
					Containers: []corev1.Container{{
						Image: "quay.io/dedgar/pod-logger:v0.0.10",
						Name:  "logger",
						SecurityContext: &corev1.SecurityContext{
							Privileged: &privileged,
							RunAsUser:  &runAsUser,
						},
						Env: []corev1.EnvVar{{
							Name:  "OO_PAUSE_ON_START",
							Value: "false",
						}, {
							Name:  "LOG_WRITER_URL",
							Value: "http://logger.openshift-scanning-operator.svc:8080/api/log",
						}, {
							Name:  "SCAN_LOG_FILE",
							Value: "/host/var/log/openshift_managed_malware_scan.log",
						}, {
							Name:  "POD_LOG_FILE",
							Value: "/host/var/log/openshift_managed_pod_creation.log",
						}},
						Ports: []corev1.ContainerPort{{
							ContainerPort: 8080,
							Name:          "logger",
						}},
						Resources: corev1.ResourceRequirements{},
						VolumeMounts: []corev1.VolumeMount{{
							Name:      "logger-secrets",
							MountPath: "/secrets",
						}, {
							Name:      "host-logs",
							MountPath: "/host/var/log/",
						}},
					}},
					Volumes: []corev1.Volume{{
						Name: "logger-secrets",
						VolumeSource: corev1.VolumeSource{
							Secret: &corev1.SecretVolumeSource{
								SecretName: "logger-secrets",
							},
						},
					}, {
						Name: "host-logs",
						VolumeSource: corev1.VolumeSource{
							HostPath: &corev1.HostPathVolumeSource{
								Path: "/var/log/",
							},
						},
					}},
				},
			},
		},
	}
	return ds
}

// LoggerService returns a new service customized for logger
func LoggerService(m *managedv1alpha1.LoggerService) *corev1.Service {
	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      m.Name,
			Namespace: m.Namespace,
			Labels: map[string]string{
				"name":    m.Name,
				"k8s-app": m.Name,
			},
		},
		Spec: corev1.ServiceSpec{
			Selector: map[string]string{
				"name": "logger",
			},
			Ports: []corev1.ServicePort{{
				Port:       8080,
				TargetPort: intstr.FromInt(8080),
				Name:       m.Name,
			}},
		},
	}
	return svc
}
