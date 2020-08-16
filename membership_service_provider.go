package fabclient

import (
	"io/ioutil"
	"os"

	"github.com/hyperledger/fabric-sdk-go/pkg/client/msp"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/context"
	mspprovider "github.com/hyperledger/fabric-sdk-go/pkg/common/providers/msp"
)

type membershipServiceProvider interface {
	createSigningIdentity(certificate, privateKey string) (mspprovider.SigningIdentity, error)
	getSigningIdentity(id string) (mspprovider.SigningIdentity, error)
}

type membershipServiceClient struct {
	client *msp.Client
}

func newMembershipServiceProvider(organization string, ctx context.ClientProvider) (membershipServiceProvider, error) {
	client, err := msp.New(ctx, msp.WithOrg(organization))
	if err != nil {
		return nil, err
	}

	var mspclient = &membershipServiceClient{
		client: client,
	}

	return mspclient, nil
}

func (m *membershipServiceClient) createSigningIdentity(certificate, privateKey string) (mspprovider.SigningIdentity, error) {
	if _, err := os.Stat(certificate); err != nil {
		return nil, err
	}

	if _, err := os.Stat(privateKey); err != nil {
		return nil, err
	}

	certificateAsBytes, err := ioutil.ReadFile(certificate)
	if err != nil {
		return nil, err
	}

	privateKeyAsBytes, err := ioutil.ReadFile(privateKey)
	if err != nil {
		return nil, err
	}

	return m.client.CreateSigningIdentity(mspprovider.WithCert(certificateAsBytes), mspprovider.WithPrivateKey(privateKeyAsBytes))
}

func (m *membershipServiceClient) getSigningIdentity(id string) (mspprovider.SigningIdentity, error) {
	return m.client.GetSigningIdentity(id)
}
