params.outDir = "./results/"

relatedness = Channel.fromPath("./example-data/relatedness-nums-and-codes.csv")
demographics = Channel.fromPath("./example-data/demographics.csv")
parentage = Channel.fromPath("./example-data/parentage.csv")
imageExts = Channel.of("svg")
reps = 10

result_path = file(params.outDir)

process relped {
    input:
        each rep from 1..reps
        file r from relatedness
        file d from demographics
        file p from parentage

    output:
        file "run-${rep}-${r}.dot" into ped

    script:
        """
        relped build \
            --relatedness ${r} \
            --demographics ${d} \
            --parentage ${p} \
            --output run-${rep}-${r}.dot
        """
}

process render {
    echo true
    input:
        each p from ped.flatMap()
        val e from imageExts

    output:
        file "${p.name}.${e}" into rendered

    publishDir "${result_path}"

    script:
        """
        dot -T${e} -o ${p.name}.${e} ${p}
        """
}
