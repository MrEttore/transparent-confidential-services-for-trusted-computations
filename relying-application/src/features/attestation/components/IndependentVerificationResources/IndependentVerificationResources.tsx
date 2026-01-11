import { Header } from './Header';
import { VerificationResource } from './VerificationResource';

export default function IndependentVerificationResources() {
    return (
        <section>
            <div className="flex flex-col gap-6 rounded-2xl bg-slate-900/60 p-6 text-slate-100">
                <Header />
                <div className="grid grid-cols-1 gap-4 md:grid-cols-2 lg:grid-cols-3">
                    <VerificationResource
                        name="Confidential workload container"
                        description="Container image that runs the confidential workload inside the TEE."
                        resourceLink="https://hub.docker.com/r/sanctuairy/llm-core/tags"
                        linkPlaceholder="View on Docker Hub"
                    />
                    <VerificationResource
                        name="Service's backend logic container"
                        description="Additional container image part of the distributed confidential computing service's backend."
                        resourceLink="https://hub.docker.com/r/sanctuairy/llm-manager/tags"
                        linkPlaceholder="View on Docker Hub"
                    />
                    <VerificationResource
                        name="Relying application container"
                        description="Container image for the UI application that users interact with."
                        resourceLink="https://hub.docker.com/r/sanctuairy/frontend/tags"
                        linkPlaceholder="View on Docker Hub"
                    />
                    <VerificationResource
                        name="Middleware container"
                        description="Container image for the API gateway or load balancer that handles requests between the relying application and the CVM."
                        resourceLink="https://hub.docker.com/r/confidential-service/api-gateway/tags"
                        linkPlaceholder="View on Docker Hub"
                    />
                    <VerificationResource
                        name="Verifier application container"
                        description="Container image that runs the Evidence Provider service to verify attestation evidence."
                        resourceLink="https://hub.docker.com/r/attestify/evidenceverifier/tags"
                        linkPlaceholder="View on Docker Hub"
                    />
                    <VerificationResource
                        name="Baseline manifest repository"
                        description="The baseline manifest used by the verifier service to audit the confidential service's TEE and infrastructure configuration."
                        resourceLink="https://anonymous.4open.science/r/transparent-confidential-services-76BB/reference-value-provider"
                        linkPlaceholder="View on GitHub"
                    />
                </div>
            </div>
        </section>
    );
}
