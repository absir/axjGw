package main

import (
	"axj/APro"
	"axj/Kt/Kt"
	"axj/Kt/KtJson"
	"context"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"reflect"
)

func main() {
	namespace := "default"
	service := "test"

	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err.Error())
	}
	// creates the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	watcher, err := clientset.CoreV1().Endpoints(namespace).Watch(context.Background(), v1.ListOptions{
		FieldSelector: fields.OneTermEqualSelector("metadata.name", service).String(),
	})

	if err != nil {
		panic(err.Error())
	}

	go func() {
		watcher.ResultChan()
		ch := watcher.ResultChan()
		for {
			c := <-ch
			reflect.ValueOf(c)
			pods, err := clientset.CoreV1().Pods(namespace).List(context.TODO(), v1.ListOptions{
				FieldSelector: fields.OneTermEqualSelector("metadata.name", service).String(),
			})

			if err != nil {
				Kt.Err(err, false)
				continue
			}

			for pod := range pods.Items {
				println(KtJson.ToJsonStr(pod))
			}
		}

	}()

	APro.Signal()
}
