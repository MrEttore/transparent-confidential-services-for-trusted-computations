import { useMutation } from '@tanstack/react-query';
import { Check, CircleX } from 'lucide-react';
import { useEffect, useState } from 'react';
import { useDispatch, useSelector } from 'react-redux';

import { AppDispatch } from '../../../../../redux/store';
import { InputField } from '../../../../../ui';
import '../../../attestationSlice';
import {
    getChallenge,
    getEvidence,
    getVerification,
    setVerificationInfrastructure,
    setVerificationQuote,
    setVerificationWorkloads,
    updateStep,
} from '../../../attestationSlice';
import {
    verifyInfrastructure,
    verifyTdxQuote,
    verifyWorkloads,
} from '../../../services/evidenceVerifierApi';
import {
    Challenge,
    Evidence,
    Infrastructure,
    Quote,
    VerificationType,
    Workloads,
} from '../../../types/attestation';
import { ModalHeader } from '../ModalHeader';

export function VerifyEvidence() {
    const [baselineManifestUrlQuote, setBaselineManifestUrlQuote] =
        useState<string>('');
    const [
        baselineManifestUrlInfrastructure,
        setBaselineManifestUrlInfrastructure,
    ] = useState<string>('');
    const [namespace, setNamespace] = useState<string>('');
    const [repository, setRepository] = useState<string>('');
    const [tag, setTag] = useState<string>('');

    const dispatch: AppDispatch = useDispatch();
    const { value: challenge } = useSelector(getChallenge) as Challenge;
    const evidence = useSelector(getEvidence) as Evidence;
    const verification = useSelector(getVerification);

    const verifyQuoteMutation = useMutation({
        mutationKey: ['verify-quote-evidence'],
        mutationFn: verifyTdxQuote,
        onSuccess: (verificationResult) => {
            dispatch(setVerificationQuote(verificationResult));
        },
        // TODO: Handle error properly
        onError: (error: Error) => {
            console.error('Error verifying TDX Quote:', error.message);
        },
    });

    const verifyWorkloadMutation = useMutation({
        mutationKey: ['verify-workload-evidence'],
        mutationFn: verifyWorkloads,
        onSuccess: (verificationResult) => {
            dispatch(setVerificationWorkloads(verificationResult));
        },
        // TODO: Handle error properly
        onError: (error: Error) => {
            console.error('Error verifying Workloads:', error.message);
        },
    });

    const verifyInfrastructureMutation = useMutation({
        mutationKey: ['verify-infrastructure-evidence'],
        mutationFn: verifyInfrastructure,
        onSuccess: (verificationResult) => {
            dispatch(setVerificationInfrastructure(verificationResult));
        },
        // TODO: Handle error properly
        onError: (error: Error) => {
            console.error('Error verifying Infrastructure:', error.message);
        },
    });

    useEffect(() => {
        if (
            verification?.quote &&
            verification?.workloads &&
            verification?.infrastructure
        )
            dispatch(updateStep({ step: 'verifyEvidence', status: 'done' }));
    }, [verification, dispatch]);

    function handleVerification(
        e: React.FormEvent<HTMLFormElement>,
        type: VerificationType,
    ) {
        e.preventDefault();

        if (type === VerificationType.QUOTE) {
            const payload = {
                issuedChallenge: challenge as string,
                baselineManifestUrl: baselineManifestUrlQuote,
                quote: evidence.quote as Quote,
            };
            verifyQuoteMutation.mutate(payload);
        }
        if (type === VerificationType.WORKLOAD) {
            const payload = {
                issuedChallenge: challenge as string,
                referenceImage: {
                    namespace,
                    repository,
                    tag,
                },
                evidence: evidence.workloads as Workloads,
            };
            verifyWorkloadMutation.mutate(payload);
        }
        if (type === VerificationType.INFRASTRUCTURE) {
            const payload = {
                issuedChallenge: challenge as string,
                baselineManifestUrl: baselineManifestUrlInfrastructure,
                evidence: evidence.infrastructure as Infrastructure,
            };
            verifyInfrastructureMutation.mutate(payload);
        }
    }

    return (
        <div className="text-slate-100">
            <div className="flex flex-col max-h-[70vh] space-y-4 overflow-y-auto pr-2">
                <ModalHeader title="Select the Evidence to Verify" />

                <div className="divide-y-1 divide-slate-700/60">
                    {/* VERIFY QUOTE */}
                    <form
                        id="verify-quote-form"
                        className="grid grid-cols-[70%_auto] gap-x-2 py-2"
                        onSubmit={(e) =>
                            handleVerification(e, VerificationType.QUOTE)
                        }
                    >
                        <div className="space-y-2">
                            <div>
                                <h3 className="text-xl font-medium">
                                    TDX Attestation Quote
                                </h3>
                                <p className="opacity-80">
                                    This verifies that the retrieved TEE
                                    attestation quote matches a trusted
                                    baseline. To do this, provide the URL to a
                                    known-good baseline manifest, which contains
                                    the expected configuration of the secure
                                    environment.
                                </p>
                            </div>
                            <div className="space-y-1">
                                <InputField
                                    label="Baseline Manifest URL"
                                    width="w-full"
                                    value={baselineManifestUrlQuote}
                                    onChange={(e) =>
                                        setBaselineManifestUrlQuote(
                                            e.target.value,
                                        )
                                    }
                                    placeholder="https://example.com/baseline-manifest-quote.jsonc"
                                />
                            </div>
                            {verifyQuoteMutation.isError && (
                                <p className="flex w-fit items-center gap-1 rounded-md bg-red-500/20 p-1 text-sm font-medium text-red-200">
                                    <CircleX size={15} />
                                    Error: {verifyQuoteMutation.error.message}
                                </p>
                            )}
                        </div>
                        <button
                            className={`mx-auto my-auto h-10 w-1/2 rounded-lg px-1 py-0.5 text-base font-semibold transition-colors duration-200 ${verifyQuoteMutation.isPending
                                    ? 'cursor-not-allowed bg-amber-500/60 text-white opacity-75'
                                    : verifyQuoteMutation.isSuccess ||
                                        verification?.quote
                                        ? 'cursor-default bg-green-500/20 text-green-200 opacity-80'
                                        : 'cursor-pointer bg-amber-500/80 text-white hover:bg-amber-500'
                                }`}
                            type="submit"
                            form="verify-quote-form"
                            disabled={
                                verifyQuoteMutation.isPending ||
                                verifyQuoteMutation.isSuccess
                            }
                        >
                            {verifyQuoteMutation.isPending ? (
                                'Verifying...'
                            ) : verifyQuoteMutation.isError ? (
                                'Retry'
                            ) : verifyQuoteMutation.isSuccess ||
                                verification?.quote ? (
                                <Check className="inline-block" />
                            ) : (
                                'Verify'
                            )}
                        </button>
                    </form>

                    {/* VERIFY WORKLOADS */}
                    <form
                        id="verify-workload-form"
                        className="grid grid-cols-[70%_auto] gap-x-2 py-2"
                        onSubmit={(e) =>
                            handleVerification(e, VerificationType.WORKLOAD)
                        }
                    >
                        <div className="space-y-2">
                            <div>
                                <h3 className="text-xl font-medium">
                                    Workload Evidence
                                </h3>
                                <p className="opacity-80">
                                    This verifies that the container images
                                    running in the backend match trusted
                                    versions. To perform the check, provide the
                                    details of the reference image (namespace,
                                    repository, and tag) you expect the service
                                    to be running.
                                </p>
                            </div>
                            <div className="flex justify-between">
                                <div>
                                    <InputField
                                        label="Namespace"
                                        width="w-full"
                                        value={namespace}
                                        onChange={(e) =>
                                            setNamespace(e.target.value)
                                        }
                                        placeholder="e.g., confidential-service"
                                    />
                                </div>
                                <div>
                                    <InputField
                                        label="Repository"
                                        width="w-full"
                                        value={repository}
                                        onChange={(e) =>
                                            setRepository(e.target.value)
                                        }
                                        placeholder="e.g., llm-core"
                                    />
                                </div>
                                <div>
                                    <InputField
                                        label="Tag"
                                        width="w-full"
                                        value={tag}
                                        onChange={(e) => setTag(e.target.value)}
                                        placeholder="e.g., latest"
                                    />
                                </div>
                            </div>
                            {verifyWorkloadMutation.isError && (
                                <p className="flex w-fit items-center gap-1 rounded-md bg-red-500/20 p-1 text-sm font-medium text-red-200">
                                    <CircleX size={15} />
                                    Error:{' '}
                                    {verifyWorkloadMutation.error.message}
                                </p>
                            )}
                        </div>
                        <button
                            className={`mx-auto my-auto h-10 w-1/2 rounded-lg px-1 py-0.5 text-base font-semibold transition-colors duration-200 ${verifyWorkloadMutation.isPending
                                    ? 'cursor-not-allowed bg-amber-500/60 text-white opacity-75'
                                    : verifyWorkloadMutation.isSuccess ||
                                        verification?.workloads
                                        ? 'cursor-default bg-green-500/20 text-green-200 opacity-80'
                                        : 'cursor-pointer bg-amber-500/80 text-white hover:bg-amber-500'
                                }`}
                            type="submit"
                            form="verify-workload-form"
                            disabled={
                                verifyWorkloadMutation.isPending ||
                                verifyWorkloadMutation.isSuccess
                            }
                        >
                            {verifyWorkloadMutation.isPending ? (
                                'Verifying...'
                            ) : verifyWorkloadMutation.isError ? (
                                'Retry'
                            ) : verifyWorkloadMutation.isSuccess ||
                                verification?.workloads ? (
                                <Check className="inline-block" />
                            ) : (
                                'Verify'
                            )}
                        </button>
                    </form>

                    {/* VERIFY INFRASTRUCTURE */}
                    <form
                        id="verify-infrastructure-form"
                        className="grid grid-cols-[70%_auto] gap-x-2 py-2"
                        onSubmit={(e) =>
                            handleVerification(
                                e,
                                VerificationType.INFRASTRUCTURE,
                            )
                        }
                    >
                        <div className="space-y-2">
                            <div>
                                <h3 className="text-xl font-medium">
                                    Infrastructure Evidence
                                </h3>
                                <p className="opacity-80">
                                    This verifies that the underlying cloud
                                    infrastructure, including the VM and boot
                                    disk, matches a trusted baseline. To do
                                    this, provide the URL to a known-good
                                    baseline manifest, which contains the
                                    expected infrastructure details.
                                </p>
                            </div>
                            <div className="space-y-1">
                                <InputField
                                    label="Baseline Manifest URL"
                                    width="w-full"
                                    value={baselineManifestUrlInfrastructure}
                                    onChange={(e) =>
                                        setBaselineManifestUrlInfrastructure(
                                            e.target.value,
                                        )
                                    }
                                    placeholder="https://anonymous.4open.science/r/transparent-confidential-services-76BB/artifacts/baseline-manifest-v2.jsonc"
                                />
                            </div>
                            {verifyInfrastructureMutation.isError && (
                                <p className="flex w-fit items-center gap-1 rounded-md bg-red-500/20 p-1 text-sm font-medium text-red-200">
                                    <CircleX size={15} />
                                    Error:{' '}
                                    {verifyInfrastructureMutation.error.message}
                                </p>
                            )}
                        </div>
                        <button
                            className={`mx-auto my-auto h-10 w-1/2 rounded-lg px-1 py-0.5 text-base font-semibold transition-colors duration-200 ${verifyInfrastructureMutation.isPending
                                    ? 'cursor-not-allowed bg-amber-500/60 text-white opacity-75'
                                    : verifyInfrastructureMutation.isSuccess ||
                                        verification?.infrastructure
                                        ? 'cursor-default bg-green-500/20 text-green-200 opacity-80'
                                        : 'cursor-pointer bg-amber-500/80 text-white hover:bg-amber-500'
                                }`}
                            type="submit"
                            form="verify-infrastructure-form"
                            disabled={
                                verifyInfrastructureMutation.isPending ||
                                verifyInfrastructureMutation.isSuccess
                            }
                        >
                            {verifyInfrastructureMutation.isPending ? (
                                'Verifying...'
                            ) : verifyInfrastructureMutation.isError ? (
                                'Retry'
                            ) : verifyInfrastructureMutation.isSuccess ||
                                verification?.infrastructure ? (
                                <Check className="inline-block" />
                            ) : (
                                'Verify'
                            )}
                        </button>
                    </form>
                </div>
            </div>
        </div>
    );
}
